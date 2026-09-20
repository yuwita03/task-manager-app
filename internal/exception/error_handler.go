package exception

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"task-manager-app/internal/model/web"
)

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %+v\n", err)

				if validationError(c, err) {
					return
				}
				if validationTypedError(c, err) { // ← BARIS BARU, taro di sini
					return
				}
				if notFoundError(c, err) {
					return
				}
				if conflictError(c, err) {
					return
				}
				if unauthorizedError(c, err) {
					return
				}
				internalServerError(c, err)
			}
		}()
		c.Next()
	}
}

func validationError(c *gin.Context, err interface{}) bool {
	exception, ok := err.(validator.ValidationErrors)
	if ok {
		c.JSON(http.StatusBadRequest, web.WebResponse{
			Code:   http.StatusBadRequest,
			Status: "BAD REQUEST",
			Data:   exception.Error(),
		})
		return true
	}
	return false
}

func notFoundError(c *gin.Context, err interface{}) bool {
	exception, ok := err.(NotFoundError)
	if ok {
		c.JSON(http.StatusNotFound, web.WebResponse{
			Code:   http.StatusNotFound,
			Status: "NOT FOUND",
			Data:   exception.Error(),
		})
		return true
	}
	return false
}

func conflictError(c *gin.Context, err interface{}) bool {
	exception, ok := err.(ConflictError)
	if ok {
		c.JSON(http.StatusConflict, web.WebResponse{
			Code:   http.StatusConflict,
			Status: "CONFLICT",
			Data:   exception.Error(),
		})
		return true
	}
	return false
}

func unauthorizedError(c *gin.Context, err interface{}) bool {
	exception, ok := err.(UnauthorizedError)
	if ok {
		c.JSON(http.StatusUnauthorized, web.WebResponse{
			Code:   http.StatusUnauthorized,
			Status: "UNAUTHORIZED",
			Data:   exception.Error(),
		})
		return true
	}
	return false
}

func internalServerError(c *gin.Context, err interface{}) {
	log.Printf("internal server error: %+v\n", err) // detail lengkap cuma di log server

	c.JSON(http.StatusInternalServerError, web.WebResponse{
		Code:   http.StatusInternalServerError,
		Status: "INTERNAL SERVER ERROR",
		Data:   "something went wrong", // client cuma dapet pesan generic
	})
}

func validationTypedError(c *gin.Context, err interface{}) bool {
	exception, ok := err.(ValidationError)
	if ok {
		c.JSON(http.StatusBadRequest, web.WebResponse{
			Code:   http.StatusBadRequest,
			Status: "BAD REQUEST",
			Data:   exception.Error(),
		})
		return true
	}
	return false
}
