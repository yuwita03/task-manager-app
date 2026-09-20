// internal/service/label_service.go
package service

import (
	"context"

	"task-manager-app/internal/exception"
	"task-manager-app/internal/model/domain"
	"task-manager-app/internal/model/web"
	"task-manager-app/internal/repository"
)

type LabelService interface {
	Create(ctx context.Context, boardId int, userId int, req web.CreateLabelRequest) web.LabelResponse
	FindByBoardId(ctx context.Context, boardId int, userId int) []web.LabelResponse
	Delete(ctx context.Context, id int, userId int)
	AttachToTask(ctx context.Context, taskId int, userId int, labelId int)
	DetachFromTask(ctx context.Context, taskId int, userId int, labelId int)
	FindByTaskId(ctx context.Context, taskId int, userId int) []web.LabelResponse
}

type labelServiceImpl struct {
	Repo       repository.LabelRepository
	TaskLabel  repository.TaskLabelRepository
	TaskRepo   repository.TaskRepository
	ListRepo   repository.ListRepository
	MemberRepo repository.BoardMemberRepository
}

func NewLabelService(repo repository.LabelRepository, taskLabel repository.TaskLabelRepository, taskRepo repository.TaskRepository, listRepo repository.ListRepository, memberRepo repository.BoardMemberRepository) LabelService {
	return &labelServiceImpl{Repo: repo, TaskLabel: taskLabel, TaskRepo: taskRepo, ListRepo: listRepo, MemberRepo: memberRepo}
}

func (s *labelServiceImpl) assertMember(ctx context.Context, boardId int, userId int) {
	if !s.MemberRepo.IsMember(ctx, boardId, userId) {
		panic(exception.NewUnauthorizedError("you are not a member of this board"))
	}
}

// helper: dapetin board_id dari taskId (lewat task -> list -> board), lalu cek membership
func (s *labelServiceImpl) assertMemberViaTask(ctx context.Context, taskId int, userId int) {
	task := s.TaskRepo.FindById(ctx, taskId)
	list := s.ListRepo.FindById(ctx, task.ListID)
	s.assertMember(ctx, list.BoardID, userId)
}

func (s *labelServiceImpl) Create(ctx context.Context, boardId int, userId int, req web.CreateLabelRequest) web.LabelResponse {
	s.assertMember(ctx, boardId, userId)

	label := domain.Label{BoardID: boardId, Name: req.Name, Color: req.Color}
	saved := s.Repo.Save(ctx, label)
	return toLabelResponse(saved)
}

func (s *labelServiceImpl) FindByBoardId(ctx context.Context, boardId int, userId int) []web.LabelResponse {
	s.assertMember(ctx, boardId, userId)

	labels := s.Repo.FindByBoardId(ctx, boardId)
	var responses []web.LabelResponse
	for _, l := range labels {
		responses = append(responses, toLabelResponse(l))
	}
	return responses
}

func (s *labelServiceImpl) Delete(ctx context.Context, id int, userId int) {
	label := s.Repo.FindById(ctx, id)
	s.assertMember(ctx, label.BoardID, userId)
	s.Repo.Delete(ctx, id)
}

func (s *labelServiceImpl) AttachToTask(ctx context.Context, taskId int, userId int, labelId int) {
	s.assertMemberViaTask(ctx, taskId, userId)
	s.Repo.FindById(ctx, labelId) // validasi label exist
	s.TaskLabel.Attach(ctx, taskId, labelId)
}

func (s *labelServiceImpl) DetachFromTask(ctx context.Context, taskId int, userId int, labelId int) {
	s.assertMemberViaTask(ctx, taskId, userId)
	s.TaskLabel.Detach(ctx, taskId, labelId)
}

func (s *labelServiceImpl) FindByTaskId(ctx context.Context, taskId int, userId int) []web.LabelResponse {
	s.assertMemberViaTask(ctx, taskId, userId)

	labels := s.TaskLabel.FindByTaskId(ctx, taskId)
	var responses []web.LabelResponse
	for _, l := range labels {
		responses = append(responses, toLabelResponse(l))
	}
	return responses
}

func toLabelResponse(label domain.Label) web.LabelResponse {
	return web.LabelResponse{ID: label.ID, BoardID: label.BoardID, Name: label.Name, Color: label.Color}
}