package database

import (
	"context"
	"librarease/internal/usecase"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Staff struct {
	ID         uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name       string          `gorm:"column:name;type:varchar(255)"`
	LibraryID  uuid.UUID       `gorm:"column:library_id;type:uuid;uniqueIndex:idx_user_library"`
	Library    *Library        `gorm:"foreignKey:LibraryID;references:ID"`
	UserID     uuid.UUID       `gorm:"column:user_id;type:uuid;uniqueIndex:idx_user_library"`
	User       *User           `gorm:"foreignKey:UserID;references:ID"`
	CreatedAt  time.Time       `gorm:"column:created_at"`
	UpdatedAt  time.Time       `gorm:"column:updated_at"`
	DeletedAt  *gorm.DeletedAt `gorm:"column:deleted_at;"`
	Borrowings []Borrowing
}

func (Staff) TableName() string {
	return "staffs"
}

func (s *service) ListStaffs(ctx context.Context, opt usecase.ListStaffsOption) ([]usecase.Staff, int, error) {
	var (
		staffs  []Staff
		ustaffs []usecase.Staff
		count   int64
	)

	db := s.db.Table("staffs").Model([]Staff{}).WithContext(ctx)

	err := db.
		Preload("Library").
		Preload("User").
		Where("library_id = ?", opt.LibraryID).
		Count(&count).
		Limit(opt.Limit).
		Offset(opt.Skip).
		Find(&staffs).
		Error

	if err != nil {
		return nil, 0, err
	}

	for _, st := range staffs {

		staff := usecase.Staff{
			ID:        st.ID,
			Name:      st.Name,
			LibraryID: st.LibraryID,
			UserID:    st.UserID,
			CreatedAt: st.CreatedAt,
			UpdatedAt: st.UpdatedAt,
			// DeleteAt:  st.DeletedAt.Time,
		}
		if st.User != nil {
			staff.User = &usecase.User{
				ID:        st.User.ID,
				Name:      st.User.Name,
				CreatedAt: st.User.CreatedAt,
				UpdatedAt: st.User.UpdatedAt,
			}
		}
		if st.Library != nil {
			staff.Library = &usecase.Library{
				ID:        st.Library.ID,
				Name:      st.Library.Name,
				CreatedAt: st.Library.CreatedAt,
				UpdatedAt: st.Library.UpdatedAt,
			}
		}
		ustaffs = append(ustaffs, staff)
	}
	return ustaffs, int(count), nil
}

func (s *service) CreateStaff(ctx context.Context, staff usecase.Staff) (usecase.Staff, error) {
	st := Staff{
		Name:      staff.Name,
		LibraryID: staff.LibraryID,
		UserID:    staff.UserID,
	}

	err := s.db.Create(&st).Preload("Library").Preload("User").Error
	if err != nil {
		return usecase.Staff{}, err
	}

	return staff, nil
}

func (s *service) GetStaffByID(ctx context.Context, id uuid.UUID) (usecase.Staff, error) {
	var st Staff
	err := s.db.
		Preload("Library").
		Preload("User").
		Where("id = ?", id).
		First(&st).
		Error
	if err != nil {
		return usecase.Staff{}, err
	}

	staff := usecase.Staff{
		ID:        st.ID,
		Name:      st.Name,
		LibraryID: st.LibraryID,
		UserID:    st.UserID,
		CreatedAt: st.CreatedAt,
		UpdatedAt: st.UpdatedAt,
		// DeleteAt:  st.DeletedAt.Time,
	}
	if st.User != nil {
		staff.User = &usecase.User{
			ID:        st.User.ID,
			Name:      st.User.Name,
			CreatedAt: st.User.CreatedAt,
			UpdatedAt: st.User.UpdatedAt,
		}
	}
	if st.Library != nil {
		staff.Library = &usecase.Library{
			ID:        st.Library.ID,
			Name:      st.Library.Name,
			CreatedAt: st.Library.CreatedAt,
			UpdatedAt: st.Library.UpdatedAt,
		}
	}
	return staff, nil
}
