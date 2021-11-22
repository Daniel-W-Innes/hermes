package hermesErrors

import (
	"github.com/gofiber/fiber/v2"
	"log"
)

type HermesError interface {
	Error() string
	LogPrivate()
	privateMessage() string
	Wrap(s string) *BaseError
}

type BaseError struct {
	fiberError     *fiber.Error
	PrivateMessage string
}

func (b *BaseError) Error() string {
	return b.fiberError.Error()
}

func (b *BaseError) LogPrivate() {
	if b.PrivateMessage != "" {
		log.Print(b.PrivateMessage)
	}
}

func (b *BaseError) privateMessage() string {
	return b.PrivateMessage
}

func (b *BaseError) Wrap(s string) *BaseError {
	b.PrivateMessage += s
	return b
}

func InternalServerError(privateMessage string) *BaseError {
	return &BaseError{
		fiberError:     fiber.ErrInternalServerError,
		PrivateMessage: privateMessage,
	}
}

func UnprocessableEntity(privateMessage string) *BaseError {
	return &BaseError{
		fiberError:     fiber.ErrUnprocessableEntity,
		PrivateMessage: privateMessage,
	}
}
