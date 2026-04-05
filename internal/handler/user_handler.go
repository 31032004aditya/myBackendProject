package handler

import (
	"backend/internal/repository"
	"backend/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userRepo repository.UserRepository
}

func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	users, err := h.userRepo.FindAll()
	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, "failed to fetch users")
		return
	}
	response.JSONSuccess(c, http.StatusOK, users)
}

type UpdateRoleRequest struct {
	Role   string `json:"role" binding:"required,oneof=viewer analyst admin"`
	Status string `json:"status" binding:"required,oneof=active inactive"`
}

func (h *UserHandler) UpdateRole(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.JSONError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSONError(c, http.StatusBadRequest, "invalid request", err.Error())
		return
	}

	err = h.userRepo.UpdateRole(uint(id), req.Role, req.Status)
	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, "failed to update user")
		return
	}

	response.JSONSuccess(c, http.StatusOK, gin.H{"message": "user updated successfully"})
}
