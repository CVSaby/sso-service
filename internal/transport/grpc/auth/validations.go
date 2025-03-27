package auth

type RegisterReqValidation struct {
	Email       string `validate:"required,email"`
	FirstName   string `validate:"required,gt=1,lte=20"`
	LastName    string `validate:"required,gt=1,lte=20"`
	PhoneNumber string `validate:"required"`
	AvatarUrl   string `validate:"required"`
	Password    string `validate:"required,gt=8,lte=16"`
}
type LoginReqValidation struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}
