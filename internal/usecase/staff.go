package usecase

import (
	"context"
	"fmt"
	"librarease/internal/config"
	"slices"
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

	role, ok := ctx.Value(config.CTX_KEY_USER_ROLE).(string)
	if !ok {
		return nil, 0, fmt.Errorf("user role not found in context")
	}
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return nil, 0, fmt.Errorf("user id not found in context")
	}

	switch role {
	case "SUPERADMIN":
		fmt.Println("[DEBUG] global superadmin")
		// ALLOW ALL
	case "ADMIN":
		fmt.Println("[DEBUG] global admin")
		// ALLOW ALL
	case "USER":
		fmt.Println("[DEBUG] global user")
		// NOTE: currently USER STAFF & ADMIN STAFF are not separated
		staffs, _, err := u.repo.ListStaffs(ctx, ListStaffsOption{
			UserID: userID.String(),
			// Using a limit of 500 for now, adjust as needed based on expected data size
			Limit: 500,
		})
		if err != nil {
			return nil, 0, err
		}
		// user is not staff
		if len(staffs) == 0 {
			fmt.Println("[DEBUG] user is not staff, filtering by user id")
			opt.UserID = userID.String()
			break
		}

		// user is staff
		fmt.Println("[DEBUG] user is staff")
		var staffLibIDs uuid.UUIDs
		for _, staff := range staffs {
			staffLibIDs = append(staffLibIDs, staff.LibraryID)
		}
		// user is staff, filtering by library ids
		if len(opt.LibraryIDs) == 0 {
			// user is staff, filters default to assigned libraries
			fmt.Println("[DEBUG] filtering by default assigned libraries")
			opt.LibraryIDs = staffLibIDs
			break
		}

		fmt.Println("[DEBUG] filtering by library ids query")
		var intersectLibIDs uuid.UUIDs
		for _, id := range opt.LibraryIDs {
			// filter out library ids that are not assigned to the staff
			if slices.Contains(staffLibIDs, id) {
				intersectLibIDs = append(intersectLibIDs, id)
			}
		}

		if len(intersectLibIDs) == 0 {
			// user is filtering by library ids but none of the ids are assigned to the staff
			fmt.Println("[DEBUG] staff filters by lib ids but none assigned")
			opt.LibraryIDs = staffLibIDs
			break
		}

		// user is filtering by library ids and some of the ids are assigned to the staff
		fmt.Println("[DEBUG] staff filters by lib ids and some assigned")
		opt.LibraryIDs = intersectLibIDs
	}
	return u.repo.ListStaffs(ctx, opt)
}

type ListStaffsOption struct {
	Skip       int
	Limit      int
	SortBy     string
	SortIn     string
	LibraryIDs uuid.UUIDs
	UserID     string
	Name       string
	StaffRole  StaffRole
}

func (u Usecase) CreateStaff(ctx context.Context, staff Staff) (Staff, error) {

	role, ok := ctx.Value(config.CTX_KEY_USER_ROLE).(string)
	if !ok {
		return Staff{}, fmt.Errorf("user role not found in context")
	}
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return Staff{}, fmt.Errorf("user id not found in context")
	}

	switch role {
	case "SUPERADMIN":
		// ALLOW
	case "ADMIN":
		// ALLlOW
	case "USER":
		staffs, _, err := u.repo.ListStaffs(ctx, ListStaffsOption{
			UserID:     userID.String(),
			LibraryIDs: uuid.UUIDs{staff.LibraryID},
		})
		if err != nil {
			return Staff{}, err
		}
		if len(staffs) == 0 {
			// TODO: implement error
			return Staff{}, fmt.Errorf("you are not staff of this library")
		}

		if staffs[0].Role != StaffRoleAdmin {
			// TODO: implement error
			return Staff{}, fmt.Errorf("you are not allowed to assign staff")
		}
	}
	st, err := u.repo.CreateStaff(ctx, staff)
	if err != nil {
		return Staff{}, err
	}
	// refresh custom claims
	err = u.refreshCustomClaims(ctx, st.UserID)
	if err != nil {
		return Staff{}, err
	}
	return st, nil
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
