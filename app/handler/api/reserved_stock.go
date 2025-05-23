package handler

import (
	"log/slog"
	"strconv"
	"warehouse-service/app/domain"
	"warehouse-service/app/handler/api/response"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type ReservedStockHandler struct {
	usecase   domain.ReservedStockUsecase
	validator *validator.Validate
}

func NewReservedStockHandler(usecase domain.ReservedStockUsecase, validator *validator.Validate) *ReservedStockHandler {
	return &ReservedStockHandler{
		usecase:   usecase,
		validator: validator,
	}
}

func (h *ReservedStockHandler) CreateReservedStock(c *fiber.Ctx) error {
	var req domain.ReservedStockCreateRequest
	if err := c.BodyParser(&req); err != nil {
		slog.ErrorContext(c.Context(), "[reservedStockHandler] CreateReservedStock", "bodyParser", err)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrBadRequest))
	}

	if err := h.validator.Struct(req); err != nil {
		slog.ErrorContext(c.Context(), "[reservedStockHandler] CreateReservedStock", "validator", err)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrValidation))
	}

	err := h.usecase.CreateReservedStock(c.Context(), req)
	if err != nil {
		slog.ErrorContext(c.Context(), "[reservedStockHandler] CreateReservedStock", "usecase", err)
		status, resp := response.FromError(err)
		return c.Status(status).JSON(resp)
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success(nil))
}

func (h *ReservedStockHandler) UpdateReservedStockStatus(c *fiber.Ctx) error {
	orderIDStr := c.Params("order_id")
	if orderIDStr == "" {
		slog.ErrorContext(c.Context(), "[reservedStockHandler] UpdateReservedStockStatus", "orderID", "missing")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrBadRequest))
	}
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil || orderID <= 0 {
		slog.ErrorContext(c.Context(), "[reservedStockHandler] UpdateReservedStockStatus", "parseInt:"+orderIDStr, err)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrBadRequest))
	}

	var req domain.ReservedStockUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		slog.ErrorContext(c.Context(), "[reservedStockHandler] UpdateReservedStockStatus", "bodyParser", err)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrBadRequest))
	}

	if err := h.validator.Struct(req); err != nil {
		slog.ErrorContext(c.Context(), "[reservedStockHandler] UpdateReservedStockStatus", "validator", err)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrValidation))
	}

	err = h.usecase.UpdateReservedStockStatusByOrderID(c.Context(), orderID, req)
	if err != nil {
		slog.ErrorContext(c.Context(), "[reservedStockHandler] UpdateReservedStockStatus", "usecase", err)
		status, resp := response.FromError(err)
		return c.Status(status).JSON(resp)
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(nil))
}
