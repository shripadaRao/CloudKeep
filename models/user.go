package models

type User struct {
	Username 	string `json:"username"`
	Password 	string `json:"password"`
	UserEmail 	string `json:"userEmail"`
	Salt 		string `json:"salt"`
	UserID 		string `json:"userId"`
}

type RegisterUser struct {
	UserID	  string 	`json:"userId"`
	Username  string    `json:"userName"`
	Password  string    `json:"password"`
	UserEmail string	`json:"userEmail"`
}

type EmailRegistrationRequest struct {
    OTP         string `json:"otp"`
    IsVerified  bool   `json:"is_verified"`
}

type VerifyRegistrationOTPModel struct {
    UserEmail string	`json:"userEmail"`
    UserId    string	`json:"userId"`
    OTP       string	`json:"otp"`
}