package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/daheige/loyalty-system/internal/application"
	"github.com/daheige/loyalty-system/internal/interfaces/response"
)

type MemberHandler struct {
	svc application.MemberService
}

func NewMemberHandler(svc application.MemberService) *MemberHandler {
	return &MemberHandler{svc: svc}
}

func (h *MemberHandler) Register(c *gin.Context) {
	var req struct {
		ShopID     string `json:"shop_id" binding:"required"`
		CustomerID string `json:"customer_id" binding:"required"`
		Email      string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	member, err := h.svc.Register(c.Request.Context(), req.ShopID, req.CustomerID, req.Email)
	if err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}

	response.Success(c, member)
}

func (h *MemberHandler) GetMember(c *gin.Context) {
	shopID := c.Query("shop_id")
	customerID := c.Query("customer_id")

	if shopID == "" || customerID == "" {
		response.BadRequest(c, "shop_id and customer_id are required")
		return
	}

	member, err := h.svc.GetMember(c.Request.Context(), shopID, customerID)
	if err != nil {
		response.NotFound(c, "member not found")
		return
	}

	response.Success(c, member)
}

func (h *MemberHandler) GetMemberByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid member id")
		return
	}

	member, err := h.svc.GetMemberByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "member not found")
		return
	}

	response.Success(c, member)
}

func (h *MemberHandler) ListMembers(c *gin.Context) {
	shopID := c.Query("shop_id")
	if shopID == "" {
		response.BadRequest(c, "shop_id is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	members, total, err := h.svc.ListMembers(c.Request.Context(), shopID, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"items":     members,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
