package database

import (
	"context"
	"time"

	"github.com/librarease/librarease/internal/usecase"

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
	Role       string          `gorm:"column:role;check:role IN ('STAFF', 'ADMIN');default:'STAFF'"`
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

	db := s.db.Model([]Staff{}).WithContext(ctx)

	if len(opt.LibraryIDs) > 0 {
		db = db.Where("library_id IN ?", opt.LibraryIDs)
	}
	if opt.UserID != "" {
		db = db.Where("user_id = ?", opt.UserID)
	}
	if opt.Name != "" {
		db = db.Where("name ILIKE ?", "%"+opt.Name+"%")
	}
	if opt.StaffRole != "" {
		db = db.Where("role = ?", opt.StaffRole)
	}

	var (
		orderIn = "DESC"
		orderBy = "created_at"
	)
	if opt.SortBy != "" {
		orderBy = opt.SortBy
	}
	if opt.SortIn != "" {
		orderIn = opt.SortIn
	}

	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := db.
		Preload("Library").
		Preload("User").
		Limit(opt.Limit).
		Offset(opt.Skip).
		Order(orderBy + " " + orderIn).
		Find(&staffs).
		Error; err != nil {

		return nil, 0, err
	}

	for _, st := range staffs {

		staff := st.ConvertToUsecase()
		if st.User != nil {
			u := st.User.ConvertToUsecase()
			staff.User = &u
		}
		if st.Library != nil {
			l := st.Library.ConvertToUsecase()
			staff.Library = &l
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
		Role:      string(staff.Role),
	}

	err := s.db.Create(&st).Error
	if err != nil {
		return usecase.Staff{}, err
	}

	return st.ConvertToUsecase(), nil
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

	staff := st.ConvertToUsecase()
	if st.User != nil {
		u := st.User.ConvertToUsecase()
		staff.User = &u
	}
	if st.Library != nil {
		l := st.Library.ConvertToUsecase()
		staff.Library = &l
	}
	return staff, nil
}

func (s *service) UpdateStaff(ctx context.Context, staff usecase.Staff) (usecase.Staff, error) {
	st := Staff{
		Name: staff.Name,
	}

	err := s.db.WithContext(ctx).Where("id = ?", staff.ID).Updates(&st).Error
	if err != nil {
		return usecase.Staff{}, err
	}

	return st.ConvertToUsecase(), nil
}

// Convert core model to Usecase
func (st Staff) ConvertToUsecase() usecase.Staff {
	var d *time.Time
	if st.DeletedAt != nil {
		d = &st.DeletedAt.Time
	}
	return usecase.Staff{
		ID:        st.ID,
		Name:      st.Name,
		LibraryID: st.LibraryID,
		UserID:    st.UserID,
		Role:      usecase.StaffRole(st.Role),
		CreatedAt: st.CreatedAt,
		UpdatedAt: st.UpdatedAt,
		DeleteAt:  d,
	}
}
