package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ==========================================
// APPROVE CODE (POST /api/admin/approve)
// Admin duyệt mã cho khách -> Đổi status từ 0 (Pending) sang 1 (Active)
// ==========================================
func ApproveCode(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Nhận mã code cần duyệt
		var input struct {
			Code string `json:"code"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu mã code"})
			return
		}

		// 2. Thực hiện Update trong Database (customers.db)
		// Chỉ update những mã đang có status = 0 (Chưa duyệt)
		// Nếu status = 1 (Đã duyệt) hoặc 2 (Đã dùng) thì không làm gì cả
		query := "UPDATE users SET status = 1 WHERE code = ? AND status = 0"
		result, err := db.Exec(query, input.Code)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi hệ thống khi duyệt mã"})
			return
		}

		// 3. Kiểm tra xem có dòng nào thực sự thay đổi không
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			// Trường hợp này xảy ra khi: Mã sai, hoặc Mã đã được duyệt rồi, hoặc Mã đã dùng xong
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Mã không tồn tại hoặc trạng thái không hợp lệ (Có thể đã duyệt rồi)",
			})
			return
		}

		// 4. Trả về thành công
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Đã kích hoạt mã thành công! Khách hàng có thể quay ngay.",
			"code":    input.Code,
		})
	}
}
func GetApprovedCustomers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Status = 1: Đã duyệt, chưa dùng
		query := "SELECT name, code, created_at FROM users WHERE status = 1 ORDER BY created_at DESC"
		rows, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy danh sách"})
			return
		}
		defer rows.Close()

		// Tạo struct tạm để chứa dữ liệu trả về
		type ApprovedUser struct {
			Name      string `json:"name"`
			Code      string `json:"code"`
			CreatedAt string `json:"created_at"` // Dùng string để đỡ bị lỗi ngày tháng
		}

		var list []ApprovedUser
		for rows.Next() {
			var u ApprovedUser
			// Scan dữ liệu từ DB
			if err := rows.Scan(&u.Name, &u.Code, &u.CreatedAt); err != nil {
				continue
			}
			list = append(list, u)
		}

		if list == nil {
			list = []ApprovedUser{}
		}

		c.JSON(http.StatusOK, gin.H{"data": list})
	}
}

// ==========================================
// DELETE CUSTOMER (POST /api/admin/delete)
// Xóa khách hàng dựa trên mã Code (Dùng cho cả Pending và Approved)
// ==========================================
func DeleteCustomer(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Code string `json:"code"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu mã code"})
			return
		}

		// Thực hiện xóa trong DB
		query := "DELETE FROM users WHERE code = ?"
		result, err := db.Exec(query, input.Code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi hệ thống khi xóa"})
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy mã này để xóa"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Đã xóa thành công!"})
	}
}

// ... (Các hàm cũ giữ nguyên)

// API: Xóa sạch dữ liệu (Reset DB) để bắt đầu ngày mới
func ResetDatabase(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Xóa tất cả user trong bảng users và history
		// (Hoặc chỉ xóa users tùy bạn, ở đây mình xóa users thôi để giữ lịch sử quay)
		_, err := db.Exec("DELETE FROM users")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi reset DB"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Đã xóa sạch dữ liệu khách hàng!"})
	}
}

// ==========================================
// RESET HISTORY (POST /api/admin/reset-history)
// Xóa sạch bảng lịch sử trúng thưởng
// ==========================================
// internal/handlers/admin.go

func ResetHistory(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Xóa sạch bảng history
		_, err := db.Exec("DELETE FROM history")
		if err != nil {
			c.JSON(500, gin.H{"error": "Lỗi xóa lịch sử"})
			return
		}

		// 2. Lệnh đặc biệt để Reset ID về 1 trong SQLite
		_, _ = db.Exec("DELETE FROM sqlite_sequence WHERE name='history'")

		c.JSON(200, gin.H{"success": true, "message": "Lịch sử đã được làm mới hoàn toàn!"})
	}
}
