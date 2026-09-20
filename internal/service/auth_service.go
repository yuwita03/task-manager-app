// internal/service/auth_service.go
package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"task-manager-app/internal/exception"
	"task-manager-app/internal/helper"
	"task-manager-app/internal/model/domain"
	"task-manager-app/internal/model/web"
	"task-manager-app/internal/repository"
)

type AuthService interface {
	Register(ctx context.Context, req web.RegisterRequest) web.AuthResponse
	Login(ctx context.Context, req web.LoginRequest) web.AuthResponse
}

type authServiceImpl struct {
	Repo        repository.UserRepository
	JWTSecret   string
	TokenExpiry time.Duration
}

func NewAuthService(repo repository.UserRepository, jwtSecret string, tokenExpiry time.Duration) AuthService {
	return &authServiceImpl{Repo: repo, JWTSecret: jwtSecret, TokenExpiry: tokenExpiry}
}

func (s *authServiceImpl) Register(ctx context.Context, req web.RegisterRequest) web.AuthResponse {
	_, found := s.Repo.FindByEmail(ctx, req.Email)
	if found {
		panic(exception.NewConflictError("email already registered"))
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	helper.PanicIfError(err)

	user := domain.User{Name: req.Name, Email: req.Email, PasswordHash: string(hashedPassword)}
	saved := s.Repo.Save(ctx, user)

	token := s.generateToken(saved.ID)

	return web.AuthResponse{
		User: web.UserResponse{
			ID:        saved.ID,
			Name:      saved.Name,
			Email:     saved.Email,
			CreatedAt: saved.CreatedAt.Format(time.RFC3339),
		},
		Token: token,
	}
}

func (s *authServiceImpl) Login(ctx context.Context, req web.LoginRequest) web.AuthResponse {
	user, found := s.Repo.FindByEmail(ctx, req.Email)
	if !found {
		panic(exception.NewUnauthorizedError("invalid credentials"))
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		panic(exception.NewUnauthorizedError("invalid credentials"))
	}

	token := s.generateToken(user.ID)

	return web.AuthResponse{
		User: web.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		},
		Token: token,
	}
}

func (s *authServiceImpl) generateToken(userID int) string {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.TokenExpiry).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.JWTSecret))
	helper.PanicIfError(err)
	return signed
}