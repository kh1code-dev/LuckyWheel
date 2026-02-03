package models

import "time"

// Struct đại diện cho 1 khách hàng
type Customer struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Status    bool      `json:"status"`
	CreatedAt time.Time `json:"create_at"`
}
