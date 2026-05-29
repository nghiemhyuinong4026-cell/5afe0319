package controllers

import (
	"net/http"

	"vehicle-management-system/config"
	"vehicle-management-system/services"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService  *services.AuthService
	auditService *services.AuditService
}

func NewAuthController(cfg *config.JWTConfig) *AuthController {
	return &AuthController{
		authService:  services.NewAuthService(cfg),
		auditService: services.NewAuditService(),
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	User  interface{} `json:"user"`
	Token string      `json:"token"`
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	user, token, err := c.authService.Login(req.Username, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Log login action
	c.auditService.Log(user.ID, "LOGIN", "User", user.ID, map[string]interface{}{
		"username": user.Username,
		"role":     user.Role,
	})

	ctx.JSON(http.StatusOK, LoginResponse{
		User:  user,
		Token: token,
	})
}
