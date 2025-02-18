package auth

type LoginValidation struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8,max=16"`
}

type RegisterValidation struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8,max=16"`
}
