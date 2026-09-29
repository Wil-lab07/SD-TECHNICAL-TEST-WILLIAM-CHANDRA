package auth

import (
	"errors"
	"log/slog"

	"github.com/gin-gonic/gin"

	"football-api/internal/modules/admin"
	"football-api/pkg/respond"
	"football-api/pkg/validation"
)

const claimsKey = "claims"

type Handler struct {
	svc Service
	log *slog.Logger
}

func NewHandler(svc Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	rg.POST("/login", h.Login)
	rg.POST("/logout", authMiddleware, h.Logout)
	rg.GET("/me", authMiddleware, h.Me)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.BadRequest(c, "invalid request body")
		return
	}
	if err := validation.Validate(&req); err != nil {
		respond.BadRequest(c, err.Error())
		return
	}

	resp, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			respond.Unauthorized(c, ErrInvalidCredentials.Error())
			return
		}
		h.log.Error("auth.Login", slog.String("error", err.Error()))
		respond.InternalError(c, "internal server error")
		return
	}
	respond.Success(c, "login successful", resp)
}

func (h *Handler) Logout(c *gin.Context) {
	claims := claimsFromContext(c)
	if claims == nil {
		respond.Unauthorized(c, ErrTokenMissing.Error())
		return
	}
	if err := h.svc.Logout(c.Request.Context(), claims); err != nil {
		h.log.Error("auth.Logout", slog.String("error", err.Error()))
		respond.InternalError(c, "internal server error")
		return
	}
	respond.Success(c, "logout successful", nil)
}

func (h *Handler) Me(c *gin.Context) {
	claims := claimsFromContext(c)
	if claims == nil {
		respond.Unauthorized(c, ErrTokenMissing.Error())
		return
	}
	resp, err := h.svc.GetMe(c.Request.Context(), claims.AdminID)
	if err != nil {
		if errors.Is(err, admin.ErrNotFound) {
			respond.NotFound(c, "admin not found")
			return
		}
		h.log.Error("auth.Me", slog.String("error", err.Error()))
		respond.InternalError(c, "internal server error")
		return
	}
	respond.Success(c, "ok", resp)
}

func claimsFromContext(c *gin.Context) *Claims {
	val, exists := c.Get(claimsKey)
	if !exists {
		return nil
	}
	claims, ok := val.(*Claims)
	if !ok {
		return nil
	}
	return claims
}
