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

type StockTransferHandler struct {
	stockTransferUsecase domain.StockTransferUsecase
	validator            *validator.Validate
}

func NewStockTransferHandler(stockTransferUsecase domain.StockTransferUsecase, validator *validator.Validate) *StockTransferHandler {
	return &StockTransferHandler{stockTransferUsecase, validator}
}

func (h *StockTransferHandler) Create(c *fiber.Ctx) error {
	var req domain.StockTransferCreateRequest
	if err := c.BodyParser(&req); err != nil {
		slog.ErrorContext(c.Context(), "[stockTransferHandler] Create", "bodyParser", err)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrBadRequest))
	}

	if err := h.validator.Struct(req); err != nil {
		slog.ErrorContext(c.Context(), "[stockTransferHandler] Create", "validation", err)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrValidation))
	}

	shopID, err := ctxutil.GetShopIDCtx(c.Context())
	if err != nil {
		slog.ErrorContext(c.Context(), "[stockTransferHandler] Create", "GetShopIDCtx", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error(domain.ErrInternal))
	}

	stockTransfer, err := h.stockTransferUsecase.CreateTransfer(c.Context(), shopID, req)
	if err != nil {
		slog.ErrorContext(c.Context(), "[stockTransferHandler] Create", "usecase", err)
		status, response := response.FromError(err)
		return c.Status(status).JSON(response)
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success(stockTransfer))
}

func (h *StockTransferHandler) GetByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		slog.ErrorContext(c.Context(), "[stockTransferHandler] GetByID", "id", "missing")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrBadRequest))
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		slog.ErrorContext(c.Context(), "[stockTransferHandler] GetByID", "parseInt:"+idStr, err)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrBadRequest))
	}

	shopIDCtx, err := ctxutil.GetShopIDCtx(c.Context())
	if err != nil {
		slog.ErrorContext(c.Context(), "[stockTransferHandler] GetByID", "GetShopIDCtx", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error(domain.ErrInternal))
	}

	var shopID *int64
	if shopIDCtx != 0 {
		shopID = &shopIDCtx
	}

	stockTransfer, err := h.stockTransferUsecase.GetTransferByID(c.Context(), id, shopID)
	if err != nil {
		slog.ErrorContext(c.Context(), "[stockTransferHandler] GetByID", "usecase", err)
		status, response := response.FromError(err)
		return c.Status(status).JSON(response)
	}
	return c.Status(fiber.StatusOK).JSON(response.Success(stockTransfer))
}

func (h *StockTransferHandler) UpdateStatus(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		slog.ErrorContext(c.Context(), "[stockTransferHandler] UpdateStatus", "id", "missing")
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrBadRequest))
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		slog.ErrorContext(c.Context(), "[stockTransferHandler] UpdateStatus", "parseInt:"+idStr, err)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrBadRequest))
	}

	var req domain.StockTransferUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		slog.ErrorContext(c.Context(), "[stockTransferHandler] UpdateStatus", "bodyParser", err)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrBadRequest))
	}

	if err := h.validator.Struct(req); err != nil {
		slog.ErrorContext(c.Context(), "[stockTransferHandler] UpdateStatus", "validation", err)
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(domain.ErrValidation))
	}

	err = h.stockTransferUsecase.UpdateTransferStatus(c.Context(), id, req)
	if err != nil {
		slog.ErrorContext(c.Context(), "[stockTransferHandler] UpdateStatus", "usecase", err)
		status, response := response.FromError(err)
		return c.Status(status).JSON(response)
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(nil))
}

func (h *StockTransferHandler) GetListStockTransfer(c *fiber.Ctx) error {
	var param domain.GetListStockTransferRequest
	if err := c.QueryParser(&param); err != nil {
		slog.WarnContext(c.Context(), "[stockTransferHandler] GetListStockTransfer", "queryParser", err)
	}

	shopID, err := ctxutil.GetShopIDCtx(c.Context())
	if err != nil {
		slog.ErrorContext(c.Context(), "[stockTransferHandler] GetListStockTransfer", "GetShopIDCtx", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error(domain.ErrInternal))
	}

	if param.Page <= 0 {
		param.Page = 1
	}
	if param.Limit <= 0 {
		param.Limit = 10
	}
	if param.Limit > 20 {
		param.Limit = 20
	}
	if param.SortBy == "" || (param.SortBy != "created_at" && param.SortBy != "id") {
		param.SortBy = "created_at"
	}
	if param.SortOrder == "" || (param.SortOrder != "asc" && param.SortOrder != "desc") {
		param.SortOrder = "desc"
	}
	if param.Status != string(domain.TransferStatusNotStarted) &&
		param.Status != string(domain.TransferStatusInProgress) &&
		param.Status != string(domain.TransferStatusCompleted) &&
		param.Status != string(domain.TransferStatusReverted) &&
		param.Status != string(domain.TransferStatusFailed) {
		param.Status = ""
	}

	stockTransfers, metadata, err := h.stockTransferUsecase.GetListStockTransfer(c.Context(), shopID, param)
	if err != nil {
		slog.ErrorContext(c.Context(), "[stockTransferHandler] GetListStockTransfer", "usecase", err)
		status, response := response.FromError(err)
		return c.Status(status).JSON(response)
	}

	return c.Status(fiber.StatusOK).JSON(response.SuccessWithMetadata(stockTransfers, metadata))
}
