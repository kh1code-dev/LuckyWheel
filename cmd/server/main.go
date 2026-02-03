package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"luckywheel/internal/database"
	"luckywheel/internal/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load .env
	if err := godotenv.Load("configs/.env"); err != nil {
		fmt.Println("⚠️ Không tìm thấy file .env, dùng mặc định.")
	}

	// 2. Kết nối Database
	database.InitDB()
	dbCus := database.CustomerDB
	dbHis := database.HistoryDB

	// 3. Setup Router
	r := gin.Default()

	// Setup CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type"}
	r.Use(cors.New(config))

	// --- CẤU HÌNH ROUTES (Chỉ khai báo 1 lần duy nhất ở đây) ---

	// A. Public (Web & Lịch sử)
	// --- CẤU HÌNH ROUTES ---

	r.Static("/web", "./static")

	// 👇 THÊM ĐOẠN NÀY 👇
	// Khi vào trang chủ (/) -> Tự động chuyển hướng sang /web/index.html
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/web/index.html")
	})
	// 👆 HẾT ĐOẠN THÊM 👆

	r.GET("/api/history", handlers.GetHistory(dbHis))
	// B. Admin (Gom nhóm lại cho gọn)
	adminGroup := r.Group("/api/admin")
	{
		// Đường dẫn thực tế sẽ là: /api/admin/pending
		adminGroup.GET("/pending", handlers.GetPendingCustomers(dbCus))
		adminGroup.GET("/approved", handlers.GetApprovedCustomers(dbCus))
		adminGroup.POST("/approve", handlers.ApproveCode(dbCus))
		adminGroup.POST("/delete", handlers.DeleteCustomer(dbCus))
		adminGroup.POST("/reset", handlers.ResetDatabase(dbCus))        // <--- Thêm dòng này
		adminGroup.POST("/reset-history", handlers.ResetHistory(dbHis)) // Reset Lịch sử
	}

	// C. Khách hàng (Có kiểm tra giờ mở cửa)
	clientGroup := r.Group("/api")
	clientGroup.Use() // Chốt chặn kiểm tra giờ
	{
		clientGroup.POST("/register", handlers.RegisterCustomer(dbCus))
		clientGroup.POST("/validate-code", handlers.ValidateCode(dbCus))
		clientGroup.POST("/spin", handlers.SpinWheel())
	}

	// 4. Chạy Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Guest: http://localhost:%s/web/index.html\n", port)
	fmt.Printf("👮 Admin: http://localhost:%s/web/admin.html\n", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Lỗi khởi động server:", err)
	}
}
