package controllers

import (
	"github.com/gin-gonic/gin"
	"modules/internal/services"
	"net/http"
)

type OwnerController struct {
	service *services.OwnerService
}

type PurchaseRequest struct {
	SpotID uint    `json:"spot_id" binding:"required"`
	Price  float64 `json:"price" binding:"required"` // 购置价格
}

func NewOwnerController(s *services.OwnerService) *OwnerController {
	return &OwnerController{service: s}
}

// PurchaseSpot 购置永久车位
// @Summary 购置永久车位
// @Description 用户购置指定的永久车位，提交车位ID和价格信息。
// @Tags owner
// @Accept json
// @Produce json
// @Example {"spot_id": 1, "price": 150000}
// @Param input body PurchaseRequest true "购置信息"
// @Security BearerAuth
// @Success 201 {object} ParkingSpotResponse "购置成功，返回车位信息"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /owner/purchase [post]
func (c *OwnerController) PurchaseSpot(ctx *gin.Context) {
	var req PurchaseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID := ctx.MustGet("userID").(uint)
	spot, err := c.service.PurchasePermanentSpot(ctx, userID, req.SpotID, req.Price)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, ToParkingSpotResponse(spot))
}
