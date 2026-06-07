package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/atilatair/realput-bg/backend/internal/middleware"
	"github.com/atilatair/realput-bg/backend/internal/model"
	"github.com/atilatair/realput-bg/backend/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// SavedHandler manages persisted guidance results per user.
type SavedHandler struct {
	store  *repository.Store
	logger *slog.Logger
}

func NewSavedHandler(store *repository.Store, logger *slog.Logger) *SavedHandler {
	return &SavedHandler{store: store, logger: logger}
}

func (h *SavedHandler) List(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized", Code: "UNAUTHORIZED"})
		return
	}

	rows, err := h.store.ListSavedResults(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("list saved results failed", "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to load saved results", Code: "DB_ERROR"})
		return
	}

	c.JSON(http.StatusOK, rows)
}

func (h *SavedHandler) Create(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized", Code: "UNAUTHORIZED"})
		return
	}

	var req model.SaveResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid request", Code: "INVALID_JSON"})
		return
	}
	if req.Query == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "query is required", Code: "VALIDATION_ERROR"})
		return
	}

	id, err := h.store.InsertSavedResult(c.Request.Context(), userID, req.Query, req.Data)
	if err != nil {
		h.logger.Error("save result failed", "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to save result", Code: "DB_ERROR"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *SavedHandler) Delete(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized", Code: "UNAUTHORIZED"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid id", Code: "VALIDATION_ERROR"})
		return
	}

	if err := h.store.DeleteSavedResult(c.Request.Context(), userID, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "not found", Code: "NOT_FOUND"})
			return
		}
		h.logger.Error("delete saved result failed", "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to delete result", Code: "DB_ERROR"})
		return
	}

	c.Status(http.StatusNoContent)
}