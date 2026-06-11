package handler

import (
	"errors"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/yourusername/user-api/internal/models"
	"github.com/yourusername/user-api/internal/service"
)

// UserHandler holds the dependencies needed by user HTTP handlers.
type UserHandler struct {
	svc      service.UserService
	log      *zap.Logger
	validate *validator.Validate
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc service.UserService, log *zap.Logger) *UserHandler {
	return &UserHandler{
		svc:      svc,
		log:      log,
		validate: validator.New(),
	}
}

// CreateUser handles POST /users.
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Could not parse request body",
			Code:    fiber.StatusBadRequest,
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "validation_error",
			Message: formatValidationErrors(err),
			Code:    fiber.StatusBadRequest,
		})
	}

	user, err := h.svc.CreateUser(c.Context(), req)
	if err != nil {
		h.log.Error("CreateUser failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create user",
			Code:    fiber.StatusInternalServerError,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

// GetUserByID handles GET /users/:id.
func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "invalid_id",
			Message: "User ID must be a positive integer",
			Code:    fiber.StatusBadRequest,
		})
	}

	user, err := h.svc.GetUserByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error:   "not_found",
				Message: "User not found",
				Code:    fiber.StatusNotFound,
			})
		}
		h.log.Error("GetUserByID failed", zap.Int32("id", id), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve user",
			Code:    fiber.StatusInternalServerError,
		})
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

// UpdateUser handles PUT /users/:id.
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "invalid_id",
			Message: "User ID must be a positive integer",
			Code:    fiber.StatusBadRequest,
		})
	}

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Could not parse request body",
			Code:    fiber.StatusBadRequest,
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "validation_error",
			Message: formatValidationErrors(err),
			Code:    fiber.StatusBadRequest,
		})
	}

	user, err := h.svc.UpdateUser(c.Context(), id, req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error:   "not_found",
				Message: "User not found",
				Code:    fiber.StatusNotFound,
			})
		}
		h.log.Error("UpdateUser failed", zap.Int32("id", id), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update user",
			Code:    fiber.StatusInternalServerError,
		})
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

// DeleteUser handles DELETE /users/:id.
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "invalid_id",
			Message: "User ID must be a positive integer",
			Code:    fiber.StatusBadRequest,
		})
	}

	if err := h.svc.DeleteUser(c.Context(), id); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error:   "not_found",
				Message: "User not found",
				Code:    fiber.StatusNotFound,
			})
		}
		h.log.Error("DeleteUser failed", zap.Int32("id", id), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete user",
			Code:    fiber.StatusInternalServerError,
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListUsers handles GET /users with optional ?page=&page_size= query params.
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 10)

	result, err := h.svc.ListUsers(c.Context(), page, pageSize)
	if err != nil {
		h.log.Error("ListUsers failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list users",
			Code:    fiber.StatusInternalServerError,
		})
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// ErrorHandler is the global Fiber error handler for unhandled errors.
func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}
	return c.Status(code).JSON(models.ErrorResponse{
		Error:   "error",
		Message: err.Error(),
		Code:    code,
	})
}

// parseID parses and validates the :id route parameter.
func parseID(c *fiber.Ctx) (int32, error) {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return int32(id), nil
}

// formatValidationErrors converts validator errors to a human-readable string.
func formatValidationErrors(err error) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		msg := ""
		for _, fe := range ve {
			msg += fe.Field() + ": " + fe.Tag() + "; "
		}
		return msg
	}
	return err.Error()
}
