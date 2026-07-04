package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/daheige/loyalty-system/internal/application"
	"github.com/daheige/loyalty-system/internal/domain/entity"
	"github.com/daheige/loyalty-system/internal/interfaces/response"
)

type PointHandler struct {
	svc application.PointService
}

func NewPointHandler(svc application.PointService) *PointHandler {
	return &PointHandler{svc: svc}
}

func (h *PointHandler) EarnPoints(c *gin.Context) {
	var req application.EarnPointsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tx, err := h.svc.EarnPoints(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, tx)
}

func (h *PointHandler) SpendPoints(c *gin.Context) {
	var req application.SpendPointsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tx, err := h.svc.SpendPoints(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, tx)
}

func (h *PointHandler) GetBalance(c *gin.Context) {
	memberID, err := strconv.ParseUint(c.Param("member_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid member id")
		return
	}

	balance, err := h.svc.GetBalance(c.Request.Context(), memberID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, balance)
}

func (h *PointHandler) GetTransactions(c *gin.Context) {
	memberID, err := strconv.ParseUint(c.Param("member_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid member id")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	txs, total, err := h.svc.GetTransactions(c.Request.Context(), memberID, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"items":     txs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *PointHandler) CalculatePoints(c *gin.Context) {
	var req struct {
		ActionType string  `json:"action_type" binding:"required"`
		Amount     float64 `json:"amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	points, err := h.svc.CalculatePoints(c.Request.Context(), entity.RuleActionType(req.ActionType), req.Amount, 1.0)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, gin.H{"points": points})
}
