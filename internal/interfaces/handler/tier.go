package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/daheige/loyalty-system/internal/application"
	"github.com/daheige/loyalty-system/internal/interfaces/response"
)

type TierHandler struct {
	svc application.TierService
}

func NewTierHandler(svc application.TierService) *TierHandler {
	return &TierHandler{svc: svc}
}

func (h *TierHandler) GetAllTiers(c *gin.Context) {
	tiers, err := h.svc.GetAllTiers(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, tiers)
}

func (h *TierHandler) GetMemberTier(c *gin.Context) {
	memberID, err := strconv.ParseUint(c.Param("member_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid member id")
		return
	}

	tier, err := h.svc.GetMemberTier(c.Request.Context(), memberID)
	if err != nil {
		response.NotFound(c, "tier not found")
		return
	}

	response.Success(c, tier)
}

func (h *TierHandler) CheckUpgrade(c *gin.Context) {
	memberID, err := strconv.ParseUint(c.Param("member_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid member id")
		return
	}

	if err := h.svc.CheckUpgrade(c.Request.Context(), memberID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "upgrade check completed"})
}
