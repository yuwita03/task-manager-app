// internal/handler/label_handler.go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"task-manager-app/internal/helper"
	"task-manager-app/internal/model/web"
	"task-manager-app/internal/service"
)

type LabelHandler struct {
	Service  service.LabelService
	Validate *validator.Validate
}

func NewLabelHandler(s service.LabelService) *LabelHandler {
	return &LabelHandler{Service: s, Validate: validator.New()}
}

func (h *LabelHandler) Create(c *gin.Context) {
	boardId, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int) // BARU

	var req web.CreateLabelRequest
	err := c.ShouldBindJSON(&req)
	helper.PanicIfError(err)
	err = h.Validate.Struct(req)
	helper.PanicIfError(err)

	result := h.Service.Create(c.Request.Context(), boardId, userId, req) // BARU: tambah userId
	c.JSON(http.StatusCreated, web.WebResponse{Code: http.StatusCreated, Status: "CREATED", Data: result})
}

func (h *LabelHandler) FindByBoardId(c *gin.Context) {
	boardId, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int) // BARU

	result := h.Service.FindByBoardId(c.Request.Context(), boardId, userId) // BARU: tambah userId
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK", Data: result})
}

func (h *LabelHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int) // BARU

	h.Service.Delete(c.Request.Context(), id, userId) // BARU: tambah userId
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK"})
}

func (h *LabelHandler) AttachToTask(c *gin.Context) {
	taskId, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int) // BARU

	var req struct {
		LabelID int `json:"label_id" validate:"required"`
	}
	err := c.ShouldBindJSON(&req)
	helper.PanicIfError(err)
	err = h.Validate.Struct(req)
	helper.PanicIfError(err)

	result := h.Service.AttachToTask(c.Request.Context(), taskId, userId, req.LabelID) // BARU: tambah userId
	c.JSON(http.StatusCreated, web.WebResponse{Code: http.StatusCreated, Status: "CREATED", Data: result})
}

func (h *LabelHandler) DetachFromTask(c *gin.Context) {
	taskId, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int) // BARU
	labelId, _ := strconv.Atoi(c.Param("labelId"))

	result := h.Service.DetachFromTask(c.Request.Context(), taskId, userId, labelId) // BARU: tambah userId
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK", Data: result})
}