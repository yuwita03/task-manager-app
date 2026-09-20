// internal/handler/board_handler.go
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

type BoardHandler struct {
	Service  service.BoardService
	Validate *validator.Validate
}

func NewBoardHandler(s service.BoardService) *BoardHandler {
	return &BoardHandler{Service: s, Validate: validator.New()}
}

func (h *BoardHandler) Create(c *gin.Context) {
	userId := c.MustGet("user_id").(int)

	var req web.CreateBoardRequest
	err := c.ShouldBindJSON(&req)
	helper.PanicIfError(err)
	err = h.Validate.Struct(req)
	helper.PanicIfError(err)

	result := h.Service.Create(c.Request.Context(), userId, req)
	c.JSON(http.StatusCreated, web.WebResponse{Code: http.StatusCreated, Status: "CREATED", Data: result})
}

func (h *BoardHandler) FindMine(c *gin.Context) {
	userId := c.MustGet("user_id").(int)
	result := h.Service.FindByUserId(c.Request.Context(), userId)
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK", Data: result})
}

func (h *BoardHandler) FindById(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int) // BARU

	result := h.Service.FindById(c.Request.Context(), id, userId) // BARU: tambah userId
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK", Data: result})
}

func (h *BoardHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int) // BARU

	var req web.UpdateBoardRequest
	err := c.ShouldBindJSON(&req)
	helper.PanicIfError(err)
	err = h.Validate.Struct(req)
	helper.PanicIfError(err)

	result := h.Service.Update(c.Request.Context(), id, userId, req) // BARU: tambah userId
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK", Data: result})
}

func (h *BoardHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int) // BARU

	h.Service.Delete(c.Request.Context(), id, userId) // BARU: tambah userId
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK"})
}

func (h *BoardHandler) InviteMember(c *gin.Context) {
	boardId, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int) // BARU

	var req web.InviteMemberRequest
	err := c.ShouldBindJSON(&req)
	helper.PanicIfError(err)
	err = h.Validate.Struct(req)
	helper.PanicIfError(err)

	result := h.Service.InviteMember(c.Request.Context(), boardId, userId, req) // BARU: tambah userId
	c.JSON(http.StatusCreated, web.WebResponse{Code: http.StatusCreated, Status: "CREATED", Data: result})
}

func (h *BoardHandler) RemoveMember(c *gin.Context) {
	boardId, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int)         // BARU: ini si pemanggil (harus owner)
	targetUserId, _ := strconv.Atoi(c.Param("userId")) // ini yang mau di-remove

	h.Service.RemoveMember(c.Request.Context(), boardId, userId, targetUserId) // BARU: urutan param
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK"})
}

func (h *BoardHandler) FindMembers(c *gin.Context) {
	boardId, _ := strconv.Atoi(c.Param("id"))
	userId := c.MustGet("user_id").(int) // BARU

	result := h.Service.FindMembers(c.Request.Context(), boardId, userId) // BARU: tambah userId
	c.JSON(http.StatusOK, web.WebResponse{Code: http.StatusOK, Status: "OK", Data: result})
}