package utils

import (
	"fmt"
	"github.com/Daniel-W-Innes/hermes/hermesErrors"
	"github.com/Daniel-W-Innes/hermes/models"
	"github.com/golang-jwt/jwt/v4"
	"strings"
)

func ValidateAuth(header string) (uint, hermesErrors.HermesError) {
	config, err := models.GetConfig()
	if err != nil {
		return 0, hermesErrors.InternalServerError(fmt.Sprintf("failed to get config: %s\n", err))
	}

	if strings.HasPrefix(header, "Bearer ") {
		token, err := jwt.Parse(strings.TrimPrefix(header, "Bearer "),
			func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
					return nil, hermesErrors.UnexpectedSigningMethod(token.Header["alg"])
				}
				return config.JWTConfig.PublicKey, nil
			},
		)
		if err != nil {
			return 0, err.(hermesErrors.HermesError)
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			return uint(claims["sub"].(float64)), nil
		} else {
			return 0, hermesErrors.NotValidToken()
		}
	} else {
		return 0, hermesErrors.MissingBearer()
	}
}
