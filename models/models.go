package models

type UserLogin struct {
	Username string `json:"username" xml:"username" form:"username" validate:"required"`
	Password string `json:"password" xml:"password" form:"password" validate:"required"`
}

type JWT struct {
	AccessToken string `json:"access_token"`
}
