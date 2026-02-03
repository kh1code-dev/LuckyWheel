package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite" // Import driver SQLite
)

// Khai báo 2 biến DB toàn cục riêng biệt
// Biến CustomerDB quản lý file customers.db
var CustomerDB *sql.DB

// Biến HistoryDB quản lý file history.db
var HistoryDB *sql.DB

func InitDB() {
	var err error

	// --- 1. Kết nối tới customers.db ---
	CustomerDB, err = sql.Open("sqlite", "customers.db")
	if err != nil {
		log.Fatal("Lỗi không thể mở customers.db:", err)
	}
	if err = CustomerDB.Ping(); err != nil {
		log.Fatal("Lỗi kết nối customers.db:", err)
	}

	// --- 2. Kết nối tới history.db ---
	HistoryDB, err = sql.Open("sqlite", "history.db")
	if err != nil {
		log.Fatal("Lỗi không thể mở history.db:", err)
	}
	if err = HistoryDB.Ping(); err != nil {
		log.Fatal("Lỗi kết nối history.db:", err)
	}

	log.Println("Đã kết nối thành công cả 2 Database!")

	// 3. Tạo bảng (Nếu file db của bạn mới tinh chưa có bảng)
	createTables()
}

func createTables() {
	// A. Tạo bảng users TRONG customers.db
	queryUsers := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		code TEXT NOT NULL UNIQUE,
		status INTEGER DEFAULT 0, 
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Lưu ý: Dùng CustomerDB.Exec
	if _, err := CustomerDB.Exec(queryUsers); err != nil {
		log.Fatal("Lỗi tạo bảng users:", err)
	}

	// B. Tạo bảng history TRONG history.db
	queryHistory := `
	CREATE TABLE IF NOT EXISTS history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		prize TEXT NOT NULL,
		won_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Lưu ý: Dùng HistoryDB.Exec
	if _, err := HistoryDB.Exec(queryHistory); err != nil {
		log.Fatal("Lỗi tạo bảng history:", err)
	}
}
