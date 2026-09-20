// internal/handler/task_handler.go
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

type TaskHandler struct {
	Service  service.TaskService
	Validate *validator.Validate
}

func NewTaskHandler(s service.TaskService) *TaskHandler {
	return &TaskHandler{Service: s, Validate: validator.New()}
}

func (h *TaskHandler) Create(c *gin.Context) {
	listId, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int)

	var req web.CreateTaskRequest
	err := c.ShouldBindJSON(&req)
	helper.PanicIfError(err)
	err = h.Validate.Struct(req)
	helper.PanicIfError(err)

	result := h.Service.Create(c.Request.Context(), listId, userId, req)
	c.JSON(http.StatusCreated, web.WebResponse{Code: http.StatusCreated, Status: "CREATED", Data: result})
}

func (h *TaskHandler) FindByListId(c *gin.Context) {
	listId, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int) // BARU

	result := h.Service.FindByListId(c.Request.Context(), listId, userId) // BARU: tambah userId
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK", Data: result})
}

func (h *TaskHandler) FindById(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int) // BARU

	result := h.Service.FindById(c.Request.Context(), id, userId) // BARU: tambah userId
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK", Data: result})
}

func (h *TaskHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int) // BARU

	var req web.UpdateTaskRequest
	err := c.ShouldBindJSON(&req)
	helper.PanicIfError(err)
	err = h.Validate.Struct(req)
	helper.PanicIfError(err)

	result := h.Service.Update(c.Request.Context(), id, userId, req) // BARU: tambah userId
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK", Data: result})
}

func (h *TaskHandler) Move(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int) // BARU

	var req web.MoveTaskRequest
	err := c.ShouldBindJSON(&req)
	helper.PanicIfError(err)
	err = h.Validate.Struct(req)
	helper.PanicIfError(err)

	result := h.Service.Move(c.Request.Context(), id, userId, req) // BARU: tambah userId
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK", Data: result})
}

func (h *TaskHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int) // BARU

	h.Service.Delete(c.Request.Context(), id, userId) // BARU: tambah userId
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK"})
}

func (h *TaskHandler) AssignUser(c *gin.Context) {
	taskId, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int) // BARU: si pemanggil

	var req struct {
		UserID int `json:"user_id" validate:"required"`
	}
	err := c.ShouldBindJSON(&req)
	helper.PanicIfError(err)
	err = h.Validate.Struct(req)
	helper.PanicIfError(err)

	h.Service.AssignUser(c.Request.Context(), taskId, userId, req.UserID) // BARU: tambah userId
	c.JSON(http.StatusCreated, web.WebResponse{Code: http.StatusCreated, Status: "CREATED"})
}

func (h *TaskHandler) UnassignUser(c *gin.Context) {
	taskId, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int)               // BARU: si pemanggil
	targetUserId, _ := strconv.Atoi(c.Param("userId")) // yang mau di-unassign

	h.Service.UnassignUser(c.Request.Context(), taskId, userId, targetUserId) // BARU: urutan param
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK"})
}

func (h *TaskHandler) FindAssignees(c *gin.Context) {
	taskId, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int) // BARU

	result := h.Service.FindAssignees(c.Request.Context(), taskId, userId) // BARU: tambah userId
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK", Data: result})
}