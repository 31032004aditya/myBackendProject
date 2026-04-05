package handler

import (
	"backend/internal/models"
	"backend/internal/repository"
	"backend/pkg/auth"
	"backend/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userRepo repository.UserRepository
}

func NewAuthHandler(userRepo repository.UserRepository) *AuthHandler {
	return &AuthHandler{userRepo: userRepo}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSONError(c, http.StatusBadRequest, "invalid request", err.Error())
		return
	}

	exists, _ := h.userRepo.FindByUsername(req.Username)
	if exists != nil {
		response.JSONError(c, http.StatusConflict, "username already exists")
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user := &models.User{
		Username: req.Username,
		Password: hashedPassword,
		Role:     "viewer", // Default role
		Status:   "active",
	}

	// First user becomes admin automatically for convenience
	count, _ := h.userRepo.FindAll()
	if len(count) == 0 {
		user.Role = "admin"
	}

	if err := h.userRepo.Create(user); err != nil {
		response.JSONError(c, http.StatusInternalServerError, "failed to create user")
		return
	}

	response.JSONSuccess(c, http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"userId":  user.ID,
		"role":    user.Role,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSONError(c, http.StatusBadRequest, "invalid request")
		return
	}

	user, err := h.userRepo.FindByUsername(req.Username)
	if err != nil || user == nil {
		response.JSONError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if user.Status != "active" {
		response.JSONError(c, http.StatusForbidden, "account is inactive")
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.Password) {
		response.JSONError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Role)
	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, "failed to generate token")
		return
	}

	response.JSONSuccess(c, http.StatusOK, gin.H{
		"token": token,
		"role":  user.Role,
	})
}
