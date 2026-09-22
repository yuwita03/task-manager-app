// internal/app/router.go
package app

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"task-manager-app/internal/config"
	"task-manager-app/internal/exception"
	"task-manager-app/internal/handler"
	"task-manager-app/internal/middleware"
	"task-manager-app/internal/repository"
	"task-manager-app/internal/service"
)

func NewRouter(db *pgxpool.Pool, cfg config.Config) *gin.Engine {
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.TokenExpiry)
	authHandler := handler.NewAuthHandler(authService)

	boardRepo := repository.NewBoardRepository(db)
	boardMemberRepo := repository.NewBoardMemberRepository(db)
	boardService := service.NewBoardService(boardRepo, boardMemberRepo, userRepo)
	boardHandler := handler.NewBoardHandler(boardService)

	listRepo := repository.NewListRepository(db)
	listService := service.NewListService(listRepo, boardMemberRepo)
	listHandler := handler.NewListHandler(listService)

	// PENTING: taskLabelRepo dideklarasikan DI SINI (sebelum taskService), bukan di bawah
	taskRepo := repository.NewTaskRepository(db)
	taskAssigneeRepo := repository.NewTaskAssigneeRepository(db)
	taskLabelRepo := repository.NewTaskLabelRepository(db)

	taskService := service.NewTaskService(taskRepo, listRepo, boardMemberRepo, taskAssigneeRepo, taskLabelRepo, userRepo) // GANTI: 6 parameter
	taskHandler := handler.NewTaskHandler(taskService)

	labelRepo := repository.NewLabelRepository(db)
	// taskLabelRepo TIDAK di-declare ulang di sini, sudah ada dari atas
	labelService := service.NewLabelService(labelRepo, taskLabelRepo, taskRepo, listRepo, boardMemberRepo)
	labelHandler := handler.NewLabelHandler(labelService)

	r := gin.Default()
	r.Use(gin.Recovery())
	r.Use(middleware.SecurityHeaders())
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))
	r.Use(exception.ErrorHandlerMiddleware())

	r.POST("/api/auth/register", middleware.RateLimiter(), authHandler.Register)
	r.POST("/api/auth/login", middleware.RateLimiter(), authHandler.Login)

	authGroup := r.Group("/api")
	authGroup.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		authGroup.GET("/auth/me", authHandler.Me)

		authGroup.POST("/boards", boardHandler.Create)
		authGroup.GET("/boards", boardHandler.FindMine)
		authGroup.GET("/boards/:id", boardHandler.FindById)
		authGroup.PUT("/boards/:id", boardHandler.Update)
		authGroup.DELETE("/boards/:id", boardHandler.Delete)

		authGroup.GET("/boards/:id/members", boardHandler.FindMembers)
		authGroup.POST("/boards/:id/members", boardHandler.InviteMember)
		authGroup.DELETE("/boards/:id/members/:userId", boardHandler.RemoveMember)

		authGroup.GET("/boards/:id/labels", labelHandler.FindByBoardId)
		authGroup.POST("/boards/:id/labels", labelHandler.Create)

		authGroup.POST("/boards/:id/lists", listHandler.Create)
		authGroup.GET("/boards/:id/lists", listHandler.FindByBoardId)
		authGroup.PUT("/lists/:id", listHandler.Update)
		authGroup.PUT("/lists/:id/reorder", listHandler.Reorder)
		authGroup.DELETE("/lists/:id", listHandler.Delete)

		authGroup.POST("/lists/:id/tasks", taskHandler.Create)
		authGroup.GET("/lists/:id/tasks", taskHandler.FindByListId)
		authGroup.GET("/tasks/:id", taskHandler.FindById)
		authGroup.PUT("/tasks/:id", taskHandler.Update)
		authGroup.PUT("/tasks/:id/move", taskHandler.Move)
		authGroup.DELETE("/tasks/:id", taskHandler.Delete)

		authGroup.GET("/tasks/:id/assignees", taskHandler.FindAssignees)
		authGroup.POST("/tasks/:id/assignees", taskHandler.AssignUser)
		authGroup.DELETE("/tasks/:id/assignees/:userId", taskHandler.UnassignUser)

		authGroup.POST("/tasks/:id/labels", labelHandler.AttachToTask)
		authGroup.DELETE("/tasks/:id/labels/:labelId", labelHandler.DetachFromTask)
		authGroup.DELETE("/labels/:id", labelHandler.Delete)
	}

	return r
}