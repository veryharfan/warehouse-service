package handler

import (
	"log/slog"
	"strconv"
	"warehouse-service/app/domain"
	"warehouse-service/app/handler/api/response"
	"warehouse-service/pkg/ctxutil"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type WarehouseHandler struct {
	warehouseUsecase domain.WarehouseService
	validator        *validator.Validate
}

func NewWarehouseHandler(warehouseUsecase domain.WarehouseService, validator *validator.Validate) *WarehouseHandler {
	return &WarehouseHandler{warehouseUsecase, validator}
}

func (h *WarehouseHandler) Create(c *fiber.Ctx) error {
	var req domain.WarehouseCreateRequest
	if err := c.BodyParser(&req); err != nil {
		slog.ErrorContext(c.Context(), "[warehouseHandler] Create", "bodyParser", err)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrBadRequest))
	}

	if err := h.validator.Struct(req); err != nil {
		slog.ErrorContext(c.Context(), "[warehouseHandler] Create", "validation", err)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrValidation))
	}

	shopID, err := ctxutil.GetShopIDCtx(c.Context())
	if err != nil {
		slog.ErrorContext(c.Context(), "[warehouseHandler] Create", "GetShopIDCtx", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error(domain.ErrInternal))
	}

	warehouse, err := h.warehouseUsecase.Create(c.Context(), shopID, &req)
	if err != nil {
		slog.ErrorContext(c.Context(), "[warehouseHandler] Create", "usecase", err)
		status, response := response.FromError(err)
		return c.Status(status).JSON(response)
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success(warehouse))
}

func (h *WarehouseHandler) GetByShopID(c *fiber.Ctx) error {
	shopIDStr := c.Params("shop_id")
	if shopIDStr == "" {
		slog.ErrorContext(c.Context(), "[warehouseHandler] GetByShopID", "shopID", "missing")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrBadRequest))
	}

	shopID, err := strconv.ParseInt(shopIDStr, 10, 64)
	if err != nil || shopID <= 0 {
		slog.ErrorContext(c.Context(), "[warehouseHandler] GetByShopID", "parseInt:"+shopIDStr, err)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrBadRequest))
	}

	warehouses, err := h.warehouseUsecase.GetByShopID(c.Context(), shopID)
	if err != nil {
		slog.ErrorContext(c.Context(), "[warehouseHandler] GetByShopID", "usecase", err)
		status, response := response.FromError(err)
		return c.Status(status).JSON(response)
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(warehouses))
}
