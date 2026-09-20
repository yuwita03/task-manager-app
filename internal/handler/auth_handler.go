// internal/handler/auth_handler.go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"task-manager-app/internal/exception"
	"task-manager-app/internal/helper"
	"task-manager-app/internal/model/web"
	"task-manager-app/internal/service"
)

type AuthHandler struct {
	Service  service.AuthService
	Validate *validator.Validate
}

func NewAuthHandler(s service.AuthService) *AuthHandler {
	return &AuthHandler{Service: s, Validate: validator.New()}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req web.RegisterRequest
	err := c.ShouldBindJSON(&req)
	helper.PanicIfError(err)

	err = h.Validate.Struct(req)
	helper.PanicIfError(err)

	result := h.Service.Register(c.Request.Context(), req)
	c.JSON(http.StatusCreated, web.WebResponse{Code: http.StatusCreated, Status: "CREATED", Data: result})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req web.LoginRequest
	err := c.ShouldBindJSON(&req)
	helper.PanicIfError(err)

	err = h.Validate.Struct(req)
	helper.PanicIfError(err)

	result := h.Service.Login(c.Request.Context(), req)
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK", Data: result})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		helper.PanicIfError(exception.NewUnauthorizedError("user not authenticated"))
	}

	id, ok := userID.(int)
	if !ok {
		helper.PanicIfError(exception.NewUnauthorizedError("invalid user session"))
	}

	user := h.Service.GetCurrentUser(c.Request.Context(), id)
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK", Data: user})
}
