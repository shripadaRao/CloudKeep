package models

type User struct {
	Username 	string `json:"username"`
	Password 	string `json:"password"`
	UserEmail 	string `json:"userEmail"`
	Salt 		string `json:"salt"`
	UserID 		string `json:"userId"`
}

type RegisterUser struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	UserEmail string `json:"userEmail"`
	OTP		  string `json:"otp"`
}

type UnVerifiedUser struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	UserEmail string `json:"userEmail"`
}

