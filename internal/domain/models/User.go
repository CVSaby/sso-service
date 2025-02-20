package models

type UserType string

const (
	SELLER   UserType = "seller"
	CUSTOMER UserType = "customer"
)

type User struct {
	ID       string
	Email    string
	PassHash []byte
	UserType UserType
}
