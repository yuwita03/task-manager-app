// internal/service/list_service.go
package service

import (
	"context"

	"task-manager-app/internal/exception"
	"task-manager-app/internal/model/domain"
	"task-manager-app/internal/model/web"
	"task-manager-app/internal/repository"
)

type ListService interface {
	Create(ctx context.Context, boardId int, userId int, req web.CreateListRequest) web.ListResponse
	FindByBoardId(ctx context.Context, boardId int, userId int) []web.ListResponse
	Update(ctx context.Context, id int, userId int, req web.UpdateListRequest) web.ListResponse
	Reorder(ctx context.Context, id int, userId int, req web.ReorderListRequest) web.ListResponse
	Delete(ctx context.Context, id int, userId int)
}

type listServiceImpl struct {
	Repo       repository.ListRepository
	MemberRepo repository.BoardMemberRepository
}

func NewListService(repo repository.ListRepository, memberRepo repository.BoardMemberRepository) ListService {
	return &listServiceImpl{Repo: repo, MemberRepo: memberRepo}
}

// helper: panggil ini sebelum operasi apapun yang butuh proteksi board
func (s *listServiceImpl) assertMember(ctx context.Context, boardId int, userId int) {
	if !s.MemberRepo.IsMember(ctx, boardId, userId) {
		panic(exception.NewUnauthorizedError("you are not a member of this board"))
	}
}

func (s *listServiceImpl) Create(ctx context.Context, boardId int, userId int, req web.CreateListRequest) web.ListResponse {
	s.assertMember(ctx, boardId, userId)

	count := s.Repo.CountByBoardId(ctx, boardId)
	list := domain.List{BoardID: boardId, Name: req.Name, Position: count}
	saved := s.Repo.Save(ctx, list)
	return toListResponse(saved)
}

func (s *listServiceImpl) FindByBoardId(ctx context.Context, boardId int, userId int) []web.ListResponse {
	s.assertMember(ctx, boardId, userId)

	lists := s.Repo.FindByBoardId(ctx, boardId)
	var responses []web.ListResponse
	for _, l := range lists {
		responses = append(responses, toListResponse(l))
	}
	return responses
}

func (s *listServiceImpl) Update(ctx context.Context, id int, userId int, req web.UpdateListRequest) web.ListResponse {
	existing := s.Repo.FindById(ctx, id) // ambil dulu buat tau BoardID-nya
	s.assertMember(ctx, existing.BoardID, userId)

	existing.Name = req.Name
	updated := s.Repo.Update(ctx, existing)
	return toListResponse(updated)
}

func (s *listServiceImpl) Reorder(ctx context.Context, id int, userId int, req web.ReorderListRequest) web.ListResponse {
	existing := s.Repo.FindById(ctx, id)
	s.assertMember(ctx, existing.BoardID, userId)

	existing.Position = req.Position
	updated := s.Repo.Update(ctx, existing)
	return toListResponse(updated)
}

func (s *listServiceImpl) Delete(ctx context.Context, id int, userId int) {
	existing := s.Repo.FindById(ctx, id) // validasi exist + dapetin BoardID sekaligus
	s.assertMember(ctx, existing.BoardID, userId)

	s.Repo.Delete(ctx, id)
}

func toListResponse(list domain.List) web.ListResponse {
	return web.ListResponse{ID: list.ID, BoardID: list.BoardID, Name: list.Name, Position: list.Position}
}