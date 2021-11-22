package utils

import (
	"fmt"
	"github.com/Daniel-W-Innes/hermes/hermesErrors"
	"github.com/Daniel-W-Innes/hermes/models"
	"github.com/golang-jwt/jwt/v4"
	"strings"
)

// ValidateAuth validate the auth header from user and get user id from jwt
func ValidateAuth(config *models.JWTConfig, header string) (uint, hermesErrors.HermesError) {
	// check if the header has the Bearer prefix
	if strings.HasPrefix(header, "Bearer ") {
		//decode token and validate signature
		token, err := jwt.Parse(strings.TrimPrefix(header, "Bearer "),
			func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
					// check alg to prevent downgrade attacks
					return nil, hermesErrors.UnexpectedSigningMethod(token.Header["alg"])
				}
				return &config.PublicKey, nil
			},
		)
		if err != nil {
			switch err.(type) {
			default:
				return 0, hermesErrors.InternalServerError(fmt.Sprintf("failed to validate auth %s\n", err))
			case hermesErrors.HermesError:
				return 0, err.(hermesErrors.HermesError)
			}
		}
		// get claims from token
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// return subject from token
			return uint(claims["sub"].(float64)), nil
		} else {
			return 0, hermesErrors.NotValidToken()
		}
	} else {
		return 0, hermesErrors.MissingBearer()
	}
}
