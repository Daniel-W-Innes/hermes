package models

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username    string
	PasswordKey []byte
	Messages    []Message `gorm:"foreignKey:OwnerID"`
}

//preHash and encode the user inputted password
func preHash(password []byte, pepperKey []byte) []byte {
	//Setup to hash the password with the pepper key as the secret
	hashedPassword := hmac.New(sha256.New, pepperKey)
	hashedPassword.Write(password)
	//Encode the resulting hash as Base64
	return []byte(base64.StdEncoding.EncodeToString(hashedPassword.Sum(nil)))
}

func (u *User) CheckPassword(passwordConfig *PasswordConfig, password []byte) error {
	return bcrypt.CompareHashAndPassword(u.PasswordKey, preHash(password, passwordConfig.PepperKey))
}

func (u *User) SetPassword(passwordConfig *PasswordConfig, password []byte) error {
	passwordKey, err := bcrypt.GenerateFromPassword(preHash(password, passwordConfig.PepperKey), passwordConfig.BcryptCost)
	if err != nil {
		return err
	}
	(*u).PasswordKey = passwordKey
	return nil
}

func (u *User) GenerateJWT(jwtConfig *JWTConfig) (*JWT, error) {
	claims := jwt.MapClaims{}
	claims["sub"] = u.ID

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	accessToken, err := token.SignedString(&jwtConfig.PrivateKey)
	if err != nil {
		return &JWT{}, err
	}
	return &JWT{AccessToken: accessToken}, nil
}
