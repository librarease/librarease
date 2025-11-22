package database

import (
	"context"
	"time"

	"github.com/librarease/librarease/internal/usecase"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type User struct {
	ID        uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name      string          `gorm:"column:name;type:varchar(255)"`
	Email     string          `gorm:"column:email;type:varchar(255)"`
	Phone     string          `gorm:"column:phone;type:varchar(255)"`
	JoinedAt  time.Time       `gorm:"column:joined_at;default:now()"`
	CreatedAt time.Time       `gorm:"column:created_at"`
	UpdatedAt time.Time       `gorm:"column:updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"column:deleted_at"`

	Staffs        []Staff
	Subscriptions []Subscription
	AuthUser      *AuthUser
	PushTokens    []PushToken
}

func (User) TableName() string {
	return "users"
}

func (s *service) ListUsers(ctx context.Context, opt usecase.ListUsersOption) ([]usecase.User, int, error) {
	var (
		users  []User
		uusers []usecase.User
		count  int64
	)

	db := s.db.Model([]User{}).WithContext(ctx)

	if opt.Name != "" {
		db = db.Where("name ILIKE ?", "%"+opt.Name+"%")
	}

	if opt.IDs != nil {
		db = db.Where("id IN ?", opt.IDs)
	}

	if opt.IncludePushTokens {
		db = db.Preload("PushTokens")
	}

	if opt.GlobalRole != "" {
		db = db.Joins("JOIN auth_users ON auth_users.user_id = users.id").Where("auth_users.global_role = ?", opt.GlobalRole)
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

	if opt.LibraryID != uuid.Nil {
		db = db.Joins("JOIN subscriptions ON subscriptions.user_id = users.id").
			Joins("JOIN memberships ON memberships.id = subscriptions.membership_id").
			Where("memberships.library_id = ?", opt.LibraryID).
			Distinct()
	}

	// Fix count for distinct users
	countDb := db
	if opt.LibraryID != uuid.Nil {
		countDb = db.Session(&gorm.Session{}) // clone db
		countDb = countDb.Distinct("users.id")
	}
	if err := countDb.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if opt.Limit > 0 {
		db = db.Limit(opt.Limit)
	}

	if opt.Skip > 0 {
		db = db.Offset(opt.Skip)
	}

	if err := db.
		Order(orderBy + " " + orderIn).
		Find(&users).
		Error; err != nil {

		return nil, 0, err
	}

	for _, u := range users {
		uu := u.ConvertToUsecase()
		for _, pt := range u.PushTokens {
			uu.PushTokens = append(uu.PushTokens, pt.ConvertToUsecase())
		}
		uusers = append(uusers, uu)
	}

	return uusers, int(count), nil
}

func (s *service) GetUserByID(ctx context.Context, id uuid.UUID, opt usecase.GetUserByIDOption) (usecase.User, error) {
	var u User

	db := s.db.WithContext(ctx).Model(&User{})

	if opt.IncludeStaffs {
		db.Preload("Staffs.Library")
	}
	db.Preload("AuthUser")

	if opt.IncludePushTokens {
		db.Preload("PushTokens")
	}

	err := db.Where("id = ?", id).First(&u).Error
	if err != nil {
		return usecase.User{}, err
	}

	uu := u.ConvertToUsecase()
	if u.AuthUser != nil {
		au := u.AuthUser.ConvertToUsecase()
		uu.AuthUser = &au
	}
	for _, st := range u.Staffs {
		ust := st.ConvertToUsecase()
		if st.Library != nil {
			l := st.Library.ConvertToUsecase()
			ust.Library = &l
		}
		uu.Staffs = append(uu.Staffs, ust)
	}
	for _, pt := range u.PushTokens {
		uu.PushTokens = append(uu.PushTokens, pt.ConvertToUsecase())
	}

	return uu, nil
}

func (s *service) CreateUser(ctx context.Context, user usecase.User) (usecase.User, error) {
	u := User{
		Name:  user.Name,
		Email: user.Email,
	}

	err := s.db.WithContext(ctx).Create(&u).Error
	if err != nil {
		return usecase.User{}, err
	}

	return u.ConvertToUsecase(), nil
}

func (s *service) UpdateUser(ctx context.Context, id uuid.UUID, user usecase.User) (usecase.User, error) {
	u := User{
		Name:  user.Name,
		Phone: user.Phone,
	}

	err := s.db.WithContext(ctx).Clauses(clause.Returning{}).Where("id = ?", id).Updates(&u).Error
	if err != nil {
		return usecase.User{}, err
	}

	if user.AuthUser != nil {
		au := AuthUser{
			UserID:     id,
			GlobalRole: user.AuthUser.GlobalRole,
		}
		err := s.db.WithContext(ctx).Where("user_id = ?", id).Updates(&au).Error
		if err != nil {
			return usecase.User{}, err
		}
	}

	return usecase.User{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Phone:     u.Phone,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}, nil
}

func (s *service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Where("id = ?", id).Delete(&User{}).Error
}

// Convert core model to Usecase
func (u User) ConvertToUsecase() usecase.User {
	var d *time.Time
	if u.DeletedAt != nil {
		d = &u.DeletedAt.Time
	}
	return usecase.User{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Phone:     u.Phone,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		DeleteAt:  d,
	}
}
