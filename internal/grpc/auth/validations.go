package auth

type RegisterReqValidation struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,gt=8,lte=16"`
}
type LoginReqValidation struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}
