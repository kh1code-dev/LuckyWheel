package handlers

import (
	"database/sql"
	"math/rand"
	"net/http"
	"time"

	"luckywheel/internal/database" // Nhớ sửa theo tên module của bạn

	"github.com/gin-gonic/gin"
)

// Cấu trúc phần thưởng (đặt ngay trong này hoặc chuyển sang models cũng được)
type Prize struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Weight int    `json:"-"` // Không gửi weight về client để tránh lộ tỉ lệ
}

// Cấu hình giải thưởng và tỉ lệ trúng
var prizes = []Prize{
	{ID: 1, Name: "Voucher 5%", Weight: 60},
	{ID: 2, Name: "Voucher 10%", Weight: 30},
	{ID: 3, Name: "Free Drink", Weight: 10},
	{ID: 4, Name: "Voucher 50%", Weight: 0}, // Tỉ lệ 0: Không bao giờ trúng
}

// ==========================================
// 1. SPIN WHEEL (POST /api/spin)
// Logic: Check Code -> Random giải -> Update Status -> Lưu History
// ==========================================
func SpinWheel() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Nhận mã code từ Client
		var input struct {
			Code string `json:"code"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu mã code"})
			return
		}

		// 2. Kiểm tra mã trong CustomerDB
		var userID int
		var userName string
		var status int

		// Query lấy thông tin khách
		queryCheck := `SELECT id, name, status FROM users WHERE code = ?`
		err := database.CustomerDB.QueryRow(queryCheck, input.Code).Scan(&userID, &userName, &status)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Mã không tồn tại!"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi hệ thống khi kiểm tra mã"})
			return
		}

		// 3. Validate Trạng thái
		if status == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Mã chưa được nhân viên kích hoạt!"})
			return
		}
		if status == 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Mã này đã được sử dụng rồi!"})
			return
		}

		// 4. LOGIC RANDOM (Weighted Random)
		totalWeight := 0
		for _, p := range prizes {
			totalWeight += p.Weight
		}

		// Seed random (để mỗi lần quay là ngẫu nhiên thật)
		rand.Seed(time.Now().UnixNano())
		rnd := rand.Intn(totalWeight) // Random từ 0 -> totalWeight - 1

		var wonPrize Prize
		currentWeight := 0
		for _, p := range prizes {
			currentWeight += p.Weight
			if rnd < currentWeight {
				wonPrize = p
				break
			}
		}

		// 5. Cập nhật Database (Transaction logic)
		// Bước A: Đánh dấu mã đã dùng (Status = 2) bên CustomerDB
		_, err = database.CustomerDB.Exec("UPDATE users SET status = 2 WHERE id = ?", userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi cập nhật trạng thái mã"})
			return
		}

		// Bước B: Lưu lịch sử trúng thưởng bên HistoryDB
		_, err = database.HistoryDB.Exec("INSERT INTO history (name, prize) VALUES (?, ?)", userName, wonPrize.Name)
		if err != nil {
			// Lưu ý: Nếu bước này lỗi thì mã đã bị trừ nhưng lịch sử chưa lưu.
			// Trong thực tế cần Transaction 2 pha (2PC) hoặc Saga, nhưng ở đây log ra là đủ.
			c.JSON(http.StatusOK, gin.H{"warning": "Đã trúng nhưng lỗi lưu lịch sử", "prize": wonPrize})
			return
		}

		// 6. Trả về kết quả cho Frontend diễn hoạt cảnh
		c.JSON(http.StatusOK, wonPrize)
	}
}

// ==========================================
// 2. GET HISTORY (GET /api/history)
// Lấy danh sách 10 người trúng giải gần nhất
// ==========================================
// internal/handlers/spin.go

func GetHistory(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sửa Query: Lấy thêm spin_at
		query := "SELECT name, prize, won_at FROM history ORDER BY id DESC LIMIT 20"
		rows, err := db.Query(query)
		if err != nil {
			c.JSON(500, gin.H{"error": "Lỗi lấy lịch sử"})
			return
		}
		defer rows.Close()

		type HistoryItem struct {
			Name  string `json:"name"`
			Prize string `json:"prize"`
			WonAt string `json:"won_at"` // Thêm trường này
		}

		var results []HistoryItem
		for rows.Next() {
			var i HistoryItem
			if err := rows.Scan(&i.Name, &i.Prize, &i.WonAt); err != nil {
				continue
			}
			results = append(results, i)
		}
		c.JSON(200, results)
	}
}
