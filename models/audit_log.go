package models

// AuditLog represents the AUDIT_LOGS table schema.
type AuditLog struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Action    string `json:"action"`
	Details   string `json:"details"`
}
