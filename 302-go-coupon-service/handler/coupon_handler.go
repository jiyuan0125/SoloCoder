package handler

import (
	"coupon-service/model"
	"coupon-service/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CouponHandler struct {
	couponService *service.CouponService
}

func NewCouponHandler() *CouponHandler {
	return &CouponHandler{
		couponService: service.NewCouponService(),
	}
}

func (h *CouponHandler) ClaimCoupon(c *gin.Context) {
	var req model.ClaimCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	if req.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户ID不能为空",
		})
		return
	}

	if req.BatchID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "批次ID无效",
		})
		return
	}

	coupon, err := h.couponService.ClaimCoupon(req.UserID, req.BatchID)
	if err != nil {
		switch err {
		case service.ErrBatchNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
		case service.ErrCouponNoStock,
			service.ErrCouponLimitExceeded:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "领券失败: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    coupon,
	})
}

func (h *CouponHandler) RedeemCoupon(c *gin.Context) {
	var req model.RedeemCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	if req.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户ID不能为空",
		})
		return
	}

	if req.CouponCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "优惠券码不能为空",
		})
		return
	}

	if req.OrderAmount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "订单金额无效",
		})
		return
	}

	result, err := h.couponService.RedeemCoupon(req.UserID, req.CouponCode, req.OrderAmount)
	if err != nil {
		switch err {
		case service.ErrCouponNotFound,
			service.ErrBatchNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
		case service.ErrCouponAlreadyRedeemed,
			service.ErrCouponExpired,
			service.ErrCouponUserMismatch,
			service.ErrCouponReturned:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "核销失败: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

func (h *CouponHandler) GetUserAvailableCoupons(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户ID不能为空",
		})
		return
	}

	coupons, err := h.couponService.GetUserAvailableCoupons(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取用户可用优惠券失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    coupons,
	})
}

func (h *CouponHandler) ReturnCoupons(c *gin.Context) {
	var req model.ReturnCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	if req.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户ID不能为空",
		})
		return
	}

	if req.BatchID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "批次ID无效",
		})
		return
	}

	returnedCount, err := h.couponService.ReturnCoupons(req.UserID, req.BatchID)
	if err != nil {
		switch err {
		case service.ErrBatchNotFound,
			service.ErrNoCouponsToReturn:
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "退回优惠券失败: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"returned_count": returnedCount,
			"message":        "成功退回" + strconv.Itoa(returnedCount) + "张优惠券",
		},
	})
}

func (h *CouponHandler) GetCouponByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "优惠券码不能为空",
		})
		return
	}

	coupon, err := h.couponService.GetCouponByCode(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取优惠券信息失败: " + err.Error(),
		})
		return
	}

	if coupon == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "优惠券不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    coupon,
	})
}
