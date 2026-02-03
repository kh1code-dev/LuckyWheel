package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"luckywheel/internal/models" // Nhớ thay 'luckywheel' bằng tên module của bạn

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ==========================================
// 1. REGISTER CUSTOMER (POST /api/register)
// Khách nhập tên -> Tạo mã Pending
// ==========================================
func RegisterCustomer(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Name string `json:"name"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
			return
		}

		if strings.TrimSpace(input.Name) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tên không được để trống"})
			return
		}

		// Tạo mã code: Lấy 8 ký tự đầu của UUID cho ngắn gọn, chuyển thành chữ hoa
		// VD: SPIN-A1B2C3D4
		randomPart := uuid.New().String()[:8]
		code := "SPIN-" + strings.ToUpper(randomPart)

		// Status = 0 (Pending - Chờ Admin duyệt)
		query := `INSERT INTO users (name, code, status) VALUES (?, ?, 0)`
		_, err := db.Exec(query, input.Name, code)

		if err != nil {
			fmt.Println("❌ LỖI DB CHI TIẾT:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi hệ thống, không thể tạo mã"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Đã gửi yêu cầu! Vui lòng chờ nhân viên xác nhận.",
		})
	}
}

// ==========================================
// 2. VALIDATE CODE (POST /api/validate-code)
// Khách nhập mã -> Kiểm tra có được quay không?
// ==========================================
func ValidateCode(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Code string `json:"code"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu mã code"})
			return
		}

		// Truy vấn Database để kiểm tra mã và trạng thái
		var name string
		var status int

		query := `SELECT name, status FROM users WHERE code = ?`
		err := db.QueryRow(query, input.Code).Scan(&name, &status)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"valid": false, "message": "Mã không tồn tại!"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi kiểm tra mã"})
			return
		}

		// Logic kiểm tra trạng thái (Status)
		// 0: Mới tạo (Chưa được Admin duyệt) -> Không được quay
		// 1: Đã duyệt -> Được quay
		// 2: Đã dùng -> Không được quay
		switch status {
		case 0:
			c.JSON(http.StatusBadRequest, gin.H{"valid": false, "message": "Mã chưa được nhân viên kích hoạt!"})
		case 2:
			c.JSON(http.StatusBadRequest, gin.H{"valid": false, "message": "Mã này đã được sử dụng rồi!"})
		case 1:
			c.JSON(http.StatusOK, gin.H{
				"valid":   true,
				"name":    name,
				"message": "Mã hợp lệ. Chúc may mắn!",
			})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"valid": false, "message": "Trạng thái mã không xác định"})
		}
	}
}

// ==========================================
// 3. GET PENDING CUSTOMERS (GET /api/admin/pending)
// Admin xem danh sách khách đang chờ để cấp mã
// ==========================================
func GetPendingCustomers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Chỉ lấy những người có status = 0 (Mới đăng ký)
		rows, err := db.Query("SELECT id, name, code, created_at FROM users WHERE status = 0 ORDER BY created_at DESC")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy danh sách"})
			return
		}
		defer rows.Close()

		var pendingList []models.Customer

		for rows.Next() {
			var cust models.Customer
			// Scan dữ liệu từ DB vào struct
			// Lưu ý: Struct Customer cần khớp với cột trong DB
			if err := rows.Scan(&cust.ID, &cust.Name, &cust.Code, &cust.CreatedAt); err != nil {
				continue
			}
			cust.Status = false // Status 0 nghĩa là chưa approve
			pendingList = append(pendingList, cust)
		}

		// Trả về mảng rỗng thay vì null nếu không có ai
		if pendingList == nil {
			pendingList = []models.Customer{}
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    pendingList,
		})
	}
}
