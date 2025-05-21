package response

import (
	"errors"
	"warehouse-service/app/domain"

	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Success  bool             `json:"success"`
	Metadata *domain.Metadata `json:"meta,omitempty"`
	Data     any              `json:"data,omitempty"`
	Error    string           `json:"error,omitempty"`
}

func Success(data any) *Response {
	return &Response{
		Success: true,
		Data:    data,
	}
}

func SuccessWithMetadata(data any, metadata domain.Metadata) *Response {
	return &Response{
		Success:  true,
		Data:     data,
		Metadata: &metadata,
	}
}

func Error(err error) *Response {
	return &Response{
		Success: false,
		Error:   err.Error(),
	}
}

func FromError(err error) (int, *Response) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		return fiber.StatusBadRequest, Error(err)
	case errors.Is(err, domain.ErrInvalidRequest):
		return fiber.StatusBadRequest, Error(err)
	case errors.Is(err, domain.ErrUnauthorized):
		return fiber.StatusUnauthorized, Error(err)
	case errors.Is(err, domain.ErrNotFound):
		return fiber.StatusNotFound, Error(err)
	case errors.Is(err, domain.ErrBadRequest):
		return fiber.StatusBadRequest, Error(err)
	default:
		return fiber.StatusInternalServerError, Error(domain.ErrInternal)
	}
}
