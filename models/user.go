package models

// User represents the USERS table schema.
type User struct {
	UID      string `json:"uid"`
	Username string `json:"username"`
	FullName string `json:"fullname"`
	Role     string `json:"role"`
	PIN      string `json:"pin"`
	IsActive int    `json:"is_active"`
}
