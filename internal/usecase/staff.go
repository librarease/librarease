package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type StaffRole string

const (
	StaffRoleAdmin StaffRole = "ADMIN"
	StaffRoleUser  StaffRole = "USER"
)

type Staff struct {
	ID        uuid.UUID
	Name      string
	LibraryID uuid.UUID
	UserID    uuid.UUID
	Role      StaffRole
	CreatedAt time.Time
	UpdatedAt time.Time
	DeleteAt  *time.Time
	User      *User
	Library   *Library
	// Borrowings []Borrowing
}

func (u Usecase) ListStaffs(ctx context.Context, opt ListStaffsOption) ([]Staff, int, error) {
	return u.repo.ListStaffs(ctx, opt)
}

type ListStaffsOption struct {
	Skip      int
	Limit     int
	SortBy    string
	SortIn    string
	LibraryID string
	UserID    string
	Name      string
	StaffRole StaffRole
}

func (u Usecase) CreateStaff(ctx context.Context, staff Staff) (Staff, error) {
	return u.repo.CreateStaff(ctx, staff)
}

func (u Usecase) GetStaffByID(ctx context.Context, id string) (Staff, error) {
	sid, err := uuid.Parse(id)
	if err != nil {
		return Staff{}, err
	}
	return u.repo.GetStaffByID(ctx, sid)
}

func (u Usecase) UpdateStaff(ctx context.Context, staff Staff) (Staff, error) {
	return u.repo.UpdateStaff(ctx, staff)
}
