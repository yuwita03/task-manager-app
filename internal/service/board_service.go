// internal/service/board_service.go
package service

import (
	"context"

	"task-manager-app/internal/exception"
	"task-manager-app/internal/model/domain"
	"task-manager-app/internal/model/web"
	"task-manager-app/internal/repository"
)

type BoardService interface {
	Create(ctx context.Context, userId int, req web.CreateBoardRequest) web.BoardResponse
	FindByUserId(ctx context.Context, userId int) []web.BoardResponse
	FindById(ctx context.Context, id int, userId int) web.BoardResponse
	Update(ctx context.Context, id int, userId int, req web.UpdateBoardRequest) web.BoardResponse
	Delete(ctx context.Context, id int, userId int)

	InviteMember(ctx context.Context, boardId int, userId int, req web.InviteMemberRequest)
	RemoveMember(ctx context.Context, boardId int, userId int, targetUserId int)
	FindMembers(ctx context.Context, boardId int, userId int) []web.BoardMemberResponse
}

type boardServiceImpl struct {
	BoardRepo  repository.BoardRepository
	MemberRepo repository.BoardMemberRepository
	UserRepo   repository.UserRepository
}

func NewBoardService(boardRepo repository.BoardRepository, memberRepo repository.BoardMemberRepository, userRepo repository.UserRepository) BoardService {
	return &boardServiceImpl{BoardRepo: boardRepo, MemberRepo: memberRepo, UserRepo: userRepo}
}

func (s *boardServiceImpl) assertMember(ctx context.Context, boardId int, userId int) {
	if !s.MemberRepo.IsMember(ctx, boardId, userId) {
		panic(exception.NewUnauthorizedError("you are not a member of this board"))
	}
}

func (s *boardServiceImpl) assertOwner(ctx context.Context, boardId int, userId int) {
	if !s.MemberRepo.IsOwner(ctx, boardId, userId) {
		panic(exception.NewUnauthorizedError("only the board owner can perform this action"))
	}
}

func (s *boardServiceImpl) Create(ctx context.Context, userId int, req web.CreateBoardRequest) web.BoardResponse {
	// tidak perlu check apapun — user baru bikin board, otomatis jadi owner
	board := domain.Board{Name: req.Name, Description: req.Description, OwnerID: userId}
	saved := s.BoardRepo.SaveWithOwner(ctx, board)
	return toBoardResponse(saved)
}

func (s *boardServiceImpl) FindByUserId(ctx context.Context, userId int) []web.BoardResponse {
	// tidak perlu check — query sudah otomatis scoped ke board milik user ini
	boards := s.BoardRepo.FindByUserId(ctx, userId)
	var responses []web.BoardResponse
	for _, b := range boards {
		responses = append(responses, toBoardResponse(b))
	}
	return responses
}

func (s *boardServiceImpl) FindById(ctx context.Context, id int, userId int) web.BoardResponse {
	s.assertMember(ctx, id, userId)
	return toBoardResponse(s.BoardRepo.FindById(ctx, id))
}

func (s *boardServiceImpl) Update(ctx context.Context, id int, userId int, req web.UpdateBoardRequest) web.BoardResponse {
	s.assertOwner(ctx, id, userId)

	existing := s.BoardRepo.FindById(ctx, id)
	existing.Name = req.Name
	existing.Description = req.Description
	updated := s.BoardRepo.Update(ctx, existing)
	return toBoardResponse(updated)
}

func (s *boardServiceImpl) Delete(ctx context.Context, id int, userId int) {
	s.assertOwner(ctx, id, userId)
	s.BoardRepo.FindById(ctx, id)
	s.BoardRepo.Delete(ctx, id)
}

func (s *boardServiceImpl) InviteMember(ctx context.Context, boardId int, userId int, req web.InviteMemberRequest) {
	s.assertOwner(ctx, boardId, userId)

	user, found := s.UserRepo.FindByEmail(ctx, req.Email)
	if !found {
		panic(exception.NewNotFoundError("user with that email not found"))
	}
	if s.MemberRepo.IsMember(ctx, boardId, user.ID) {
		panic(exception.NewConflictError("user is already a member of this board"))
	}
	s.MemberRepo.AddMember(ctx, boardId, user.ID, "member")
}

func (s *boardServiceImpl) RemoveMember(ctx context.Context, boardId int, userId int, targetUserId int) {
	s.assertOwner(ctx, boardId, userId)
	s.MemberRepo.RemoveMember(ctx, boardId, targetUserId)
}

func (s *boardServiceImpl) FindMembers(ctx context.Context, boardId int, userId int) []web.BoardMemberResponse {
	s.assertMember(ctx, boardId, userId)
	return s.MemberRepo.FindMembers(ctx, boardId)
}

func toBoardResponse(board domain.Board) web.BoardResponse {
	return web.BoardResponse{ID: board.ID, Name: board.Name, Description: board.Description, OwnerID: board.OwnerID}
}