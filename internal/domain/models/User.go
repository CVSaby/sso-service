package models

type UserType string

const (
	USER  UserType = "user"
	ADMIN UserType = "admin"
)

type User struct {
	ID       string
	Email    string
	PassHash []byte
	UserType UserType
}
