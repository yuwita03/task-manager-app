// internal/repository/board_member_repository.go
package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"task-manager-app/internal/helper"
	"task-manager-app/internal/model/web"
)

type BoardMemberRepository interface {
	AddMember(ctx context.Context, boardId int, userId int, role string)
	RemoveMember(ctx context.Context, boardId int, userId int)
	FindMembers(ctx context.Context, boardId int) []web.BoardMemberResponse
	IsMember(ctx context.Context, boardId int, userId int) bool
	IsOwner(ctx context.Context, boardId int, userId int) bool
}

type boardMemberRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewBoardMemberRepository(db *pgxpool.Pool) BoardMemberRepository {
	return &boardMemberRepositoryImpl{DB: db}
}

func (r *boardMemberRepositoryImpl) IsOwner(ctx context.Context, boardId int, userId int) bool {
	var role string
	query := `SELECT role FROM board_members WHERE board_id = $1 AND user_id = $2`
	err := r.DB.QueryRow(ctx, query, boardId, userId).Scan(&role)
	if err != nil {
		return false
	}
	return role == "owner"
}

func (r *boardMemberRepositoryImpl) AddMember(ctx context.Context, boardId int, userId int, role string) {
	query := `INSERT INTO board_members (board_id, user_id, role) VALUES ($1, $2, $3)`
	_, err := r.DB.Exec(ctx, query, boardId, userId, role)
	helper.PanicIfError(err)
}

func (r *boardMemberRepositoryImpl) RemoveMember(ctx context.Context, boardId int, userId int) {
	query := `DELETE FROM board_members WHERE board_id = $1 AND user_id = $2`
	_, err := r.DB.Exec(ctx, query, boardId, userId)
	helper.PanicIfError(err)
}

func (r *boardMemberRepositoryImpl) FindMembers(ctx context.Context, boardId int) []web.BoardMemberResponse {
	query := `
		SELECT u.id, u.name, u.email, bm.role
		FROM board_members bm
		JOIN users u ON u.id = bm.user_id
		WHERE bm.board_id = $1`
	rows, err := r.DB.Query(ctx, query, boardId)
	helper.PanicIfError(err)
	defer rows.Close()

	var members []web.BoardMemberResponse
	for rows.Next() {
		var m web.BoardMemberResponse
		err := rows.Scan(&m.UserID, &m.Name, &m.Email, &m.Role)
		helper.PanicIfError(err)
		members = append(members, m)
	}
	return members
}

func (r *boardMemberRepositoryImpl) IsMember(ctx context.Context, boardId int, userId int) bool {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM board_members WHERE board_id = $1 AND user_id = $2)`
	err := r.DB.QueryRow(ctx, query, boardId, userId).Scan(&exists)
	helper.PanicIfError(err)
	return exists
}