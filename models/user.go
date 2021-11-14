package models

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID          int    `db:"id"`
	Username    string `db:"username"`
	PasswordKey []byte `db:"password_key"`
}

//preHash and encode the user inputted password
func preHash(password []byte, pepperKey []byte) []byte {
	//Setup to hash the password with the pepper key as the secret
	hashedPassword := hmac.New(sha256.New, pepperKey)
	hashedPassword.Write(password)
	//Encode the resulting hash as Base64
	return []byte(base64.StdEncoding.EncodeToString(hashedPassword.Sum(nil)))
}

func (u *User) CheckPassword(password []byte) error {
	return bcrypt.CompareHashAndPassword(u.PasswordKey, preHash(password, config.PasswordConfig.PepperKey))
}

func (u *User) SetPassword(password []byte) error {
	passwordKey, err := bcrypt.GenerateFromPassword(preHash(password, config.PasswordConfig.PepperKey), config.PasswordConfig.BcryptCost)
	if err != nil {
		return err
	}
	(*u).PasswordKey = passwordKey
	return nil
}

func (u *User) Insert(db *sqlx.DB) error {
	row := db.QueryRow("INSERT INTO app_user (username, password_key) VALUES ($1,$2) RETURNING id", u.Username, u.PasswordKey)
	err := row.Scan(&u.ID)
	return err
}

func (u *User) Get(db *sqlx.DB) error {
	return db.Get(u, "SELECT * FROM app_user WHERE username=$1;", u.Username)

}

func (u *User) GenerateJWT() (*JWT, error) {
	claims := jwt.MapClaims{}
	claims["sub"] = u.ID

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	accessToken, err := token.SignedString(config.JWTConfig.PrivateKey)
	if err != nil {
		return &JWT{}, err
	}
	return &JWT{AccessToken: accessToken}, nil
}
