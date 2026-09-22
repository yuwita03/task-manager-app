// internal/service/task_service.go
package service

import (
	"context"
	"time"

	"task-manager-app/internal/exception"
	"task-manager-app/internal/model/domain"
	"task-manager-app/internal/model/web"
	"task-manager-app/internal/repository"
)

type TaskService interface {
	Create(ctx context.Context, listId int, userId int, req web.CreateTaskRequest) web.TaskResponse
	FindByListId(ctx context.Context, listId int, userId int) []web.TaskResponse
	FindById(ctx context.Context, id int, userId int) web.TaskResponse
	Update(ctx context.Context, id int, userId int, req web.UpdateTaskRequest) web.TaskResponse
	Move(ctx context.Context, id int, userId int, req web.MoveTaskRequest) web.TaskResponse
	Delete(ctx context.Context, id int, userId int)

	AssignUser(ctx context.Context, taskId int, userId int, targetUserId int) []web.UserResponse
	UnassignUser(ctx context.Context, taskId int, userId int, targetUserId int) []web.UserResponse
	FindAssignees(ctx context.Context, taskId int, userId int) []web.UserResponse
}

type taskServiceImpl struct {
	Repo         repository.TaskRepository
	ListRepo     repository.ListRepository
	MemberRepo   repository.BoardMemberRepository
	AssigneeRepo repository.TaskAssigneeRepository
	LabelRepo    repository.TaskLabelRepository
	UserRepo     repository.UserRepository
}

func NewTaskService(repo repository.TaskRepository, listRepo repository.ListRepository, memberRepo repository.BoardMemberRepository, assigneeRepo repository.TaskAssigneeRepository, labelRepo repository.TaskLabelRepository, userRepo repository.UserRepository) TaskService {
	return &taskServiceImpl{Repo: repo, ListRepo: listRepo, MemberRepo: memberRepo, AssigneeRepo: assigneeRepo, LabelRepo: labelRepo, UserRepo: userRepo}
}

func (s *taskServiceImpl) assertMemberViaList(ctx context.Context, listId int, userId int) {
	list := s.ListRepo.FindById(ctx, listId)
	if !s.MemberRepo.IsMember(ctx, list.BoardID, userId) {
		panic(exception.NewUnauthorizedError("you are not a member of this board"))
	}
}

func (s *taskServiceImpl) assertMemberViaTask(ctx context.Context, taskId int, userId int) domain.Task {
	task := s.Repo.FindById(ctx, taskId)
	s.assertMemberViaList(ctx, task.ListID, userId)
	return task
}

func parseDueDate(strPtr *string) *time.Time {
	if strPtr == nil || *strPtr == "" {
		return nil
	}
	if t, err := time.Parse("2006-01-02", *strPtr); err == nil {
		return &t
	}
	if t, err := time.Parse(time.RFC3339, *strPtr); err == nil {
		return &t
	}
	panic(exception.NewValidationError("due_date must be YYYY-MM-DD or RFC3339 format"))
}

func (s *taskServiceImpl) Create(ctx context.Context, listId int, userId int, req web.CreateTaskRequest) web.TaskResponse {
	s.assertMemberViaList(ctx, listId, userId)

	count := s.Repo.CountByListId(ctx, listId)
	task := domain.Task{
		ListID: listId, Title: req.Title, Description: req.Description,
		Status: "todo", Position: count, DueDate: parseDueDate(req.DueDate), CreatedBy: userId,
	}
	saved := s.Repo.Save(ctx, task)
	return s.toTaskResponse(ctx, saved)
}

func (s *taskServiceImpl) FindByListId(ctx context.Context, listId int, userId int) []web.TaskResponse {
	s.assertMemberViaList(ctx, listId, userId)

	tasks := s.Repo.FindByListId(ctx, listId)
	var responses []web.TaskResponse
	for _, t := range tasks {
		responses = append(responses, s.toTaskResponse(ctx, t))
	}
	return responses
}

func (s *taskServiceImpl) FindById(ctx context.Context, id int, userId int) web.TaskResponse {
	task := s.assertMemberViaTask(ctx, id, userId)
	return s.toTaskResponse(ctx, task)
}

func (s *taskServiceImpl) Update(ctx context.Context, id int, userId int, req web.UpdateTaskRequest) web.TaskResponse {
	existing := s.assertMemberViaTask(ctx, id, userId)

	existing.Title = req.Title
	existing.Description = req.Description
	existing.Status = req.Status
	existing.DueDate = parseDueDate(req.DueDate)
	updated := s.Repo.Update(ctx, existing)
	return s.toTaskResponse(ctx, updated)
}

func (s *taskServiceImpl) Move(ctx context.Context, id int, userId int, req web.MoveTaskRequest) web.TaskResponse {
	s.assertMemberViaTask(ctx, id, userId)
	s.assertMemberViaList(ctx, req.ListID, userId)

	moved := s.Repo.Move(ctx, id, req.ListID, req.Position)
	return s.toTaskResponse(ctx, moved)
}

func (s *taskServiceImpl) Delete(ctx context.Context, id int, userId int) {
	s.assertMemberViaTask(ctx, id, userId)
	s.Repo.Delete(ctx, id)
}

func (s *taskServiceImpl) AssignUser(ctx context.Context, taskId int, userId int, targetUserId int) []web.UserResponse {
	s.assertMemberViaTask(ctx, taskId, userId)
	s.UserRepo.FindById(ctx, targetUserId)
	s.AssigneeRepo.Assign(ctx, taskId, targetUserId)
	return s.AssigneeRepo.FindByTaskId(ctx, taskId)
}

func (s *taskServiceImpl) UnassignUser(ctx context.Context, taskId int, userId int, targetUserId int) []web.UserResponse {
	s.assertMemberViaTask(ctx, taskId, userId)
	s.AssigneeRepo.Unassign(ctx, taskId, targetUserId)
	return s.AssigneeRepo.FindByTaskId(ctx, taskId)
}

func (s *taskServiceImpl) FindAssignees(ctx context.Context, taskId int, userId int) []web.UserResponse {
	s.assertMemberViaTask(ctx, taskId, userId)
	return s.AssigneeRepo.FindByTaskId(ctx, taskId)
}

func (s *taskServiceImpl) toTaskResponse(ctx context.Context, task domain.Task) web.TaskResponse {
	resp := web.TaskResponse{
		ID: task.ID, ListID: task.ListID, Title: task.Title, Description: task.Description,
		Status: task.Status, Position: task.Position, CreatedBy: task.CreatedBy,
	}
	if task.DueDate != nil {
		resp.DueDate = task.DueDate.Format(time.RFC3339)
	}

	resp.Assignees = s.AssigneeRepo.FindByTaskId(ctx, task.ID)

	labels := s.LabelRepo.FindByTaskId(ctx, task.ID)
	var labelResponses []web.LabelResponse
	for _, l := range labels {
		labelResponses = append(labelResponses, web.LabelResponse{ID: l.ID, BoardID: l.BoardID, Name: l.Name, Color: l.Color})
	}
	resp.Labels = labelResponses

	return resp
}