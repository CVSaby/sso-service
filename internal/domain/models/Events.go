package models

type UserRegisteredEvent struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	AvatarUrl   string `json:"avatar_url"`
	TraceID     string `json:"trace_id"`
}
