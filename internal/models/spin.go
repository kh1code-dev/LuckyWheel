package models

// Struct đại diện cho 1 lần quay
type SpinHistory struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Prize string `json:"prize"`
	WonAt string `json:"won_at"`
}
