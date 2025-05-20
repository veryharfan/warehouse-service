package handler

import (
	"log/slog"
	"strconv"
	"warehouse-service/app/domain"
	"warehouse-service/app/handler/api/response"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type StockHandler struct {
	stockUsecase domain.StockService
	validator    *validator.Validate
}

func NewStockHandler(stockUsecase domain.StockService, validator *validator.Validate) *StockHandler {
	return &StockHandler{
		stockUsecase: stockUsecase,
		validator:    validator,
	}
}

func (h *StockHandler) Create(c *fiber.Ctx) error {
	var req domain.StockCreateRequest
	if err := c.BodyParser(&req); err != nil {
		slog.ErrorContext(c.Context(), "[stockHandler] Create", "bodyParser", err)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrBadRequest))
	}

	if err := h.validator.Struct(req); err != nil {
		slog.ErrorContext(c.Context(), "[stockHandler] Create", "validation", err)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrValidation))
	}

	stock, err := h.stockUsecase.InitStock(c.Context(), req)
	if err != nil {
		slog.ErrorContext(c.Context(), "[stockHandler] Create", "usecase", err)
		status, resp := response.FromError(err)
		return c.Status(status).JSON(resp)
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success(stock))
}

func (h *StockHandler) GetByProductID(c *fiber.Ctx) error {
	productIDStr := c.Params("product_id")
	if productIDStr == "" {
		slog.ErrorContext(c.Context(), "[stockHandler] GetByProductID", "productID", "missing")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrBadRequest))
	}

	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil || productID <= 0 {
		slog.ErrorContext(c.Context(), "[stockHandler] GetByProductID", "parseInt:"+productIDStr, err)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrBadRequest))
	}

	stocks, err := h.stockUsecase.GetByProductID(c.Context(), productID)
	if err != nil {
		slog.ErrorContext(c.Context(), "[stockHandler] GetByProductID", "usecase", err)
		status, resp := response.FromError(err)
		return c.Status(status).JSON(resp)
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(stocks))
}

func (h *StockHandler) UpdateQuantity(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		slog.ErrorContext(c.Context(), "[stockHandler] UpdateQuantity", "id", "missing")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrBadRequest))
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		slog.ErrorContext(c.Context(), "[stockHandler] UpdateQuantity", "parseInt:"+idStr, err)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrBadRequest))
	}

	var req domain.UpdateQuantityRequest
	if err := c.BodyParser(&req); err != nil {
		slog.ErrorContext(c.Context(), "[stockHandler] UpdateQuantity", "bodyParser", err)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrBadRequest))
	}

	err = h.stockUsecase.UpdateQuantity(c.Context(), id, req.Quantity)
	if err != nil {
		slog.ErrorContext(c.Context(), "[stockHandler] UpdateQuantity", "usecase", err)
		status, resp := response.FromError(err)
		return c.Status(status).JSON(resp)
	}

	return c.Status(fiber.StatusNoContent).JSON(response.Success(nil))
}
