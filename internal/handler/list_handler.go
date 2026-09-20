// internal/handler/list_handler.go
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

type ListHandler struct {
	Service  service.ListService
	Validate *validator.Validate
}

func NewListHandler(s service.ListService) *ListHandler {
	return &ListHandler{Service: s, Validate: validator.New()}
}

func (h *ListHandler) Create(c *gin.Context) {
	boardId, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int)

	var req web.CreateListRequest
	err := c.ShouldBindJSON(&req)
	helper.PanicIfError(err)
	err = h.Validate.Struct(req)
	helper.PanicIfError(err)

	result := h.Service.Create(c.Request.Context(), boardId, userId, req)
	c.JSON(http.StatusCreated, web.WebResponse{Code: http.StatusCreated, Status: "CREATED", Data: result})
}

func (h *ListHandler) FindByBoardId(c *gin.Context) {
	boardId, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int)

	result := h.Service.FindByBoardId(c.Request.Context(), boardId, userId)
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK", Data: result})
}

func (h *ListHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int)

	var req web.UpdateListRequest
	err := c.ShouldBindJSON(&req)
	helper.PanicIfError(err)
	err = h.Validate.Struct(req)
	helper.PanicIfError(err)

	result := h.Service.Update(c.Request.Context(), id, userId, req)
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK", Data: result})
}

func (h *ListHandler) Reorder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int)

	var req web.ReorderListRequest
	err := c.ShouldBindJSON(&req)
	helper.PanicIfError(err)
	err = h.Validate.Struct(req)
	helper.PanicIfError(err)

	result := h.Service.Reorder(c.Request.Context(), id, userId, req)
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK", Data: result})
}

func (h *ListHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int)

	h.Service.Delete(c.Request.Context(), id, userId)
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK"})
}