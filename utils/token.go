package utils

import (
	"fmt"
	"github.com/Daniel-W-Innes/hermes/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"log"
	"strings"
)

func ValidateAuth(header string) (uint, error) {
	config, err := models.GetConfig()
	if err != nil {
		log.Printf("failed to get config: %s\n", err)
		return 0, fiber.ErrInternalServerError
	}

	if strings.HasPrefix(header, "Bearer ") {
		token, err := jwt.Parse(strings.TrimPrefix(header, "Bearer "),
			func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
					return nil, fiber.NewError(fiber.StatusUnauthorized, fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]))
				}
				return config.JWTConfig.PublicKey, nil
			},
		)
		if err != nil {
			return 0, err
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			return uint(claims["sub"].(float64)), nil
		} else {
			return 0, fiber.NewError(fiber.StatusUnauthorized, "token is not valid")
		}
	} else {
		return 0, fiber.NewError(fiber.StatusUnauthorized, "missing bearer in header")
	}
}
