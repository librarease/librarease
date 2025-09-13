package database

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/librarease/librarease/internal/usecase"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PushToken struct {
	ID        uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID    uuid.UUID       `gorm:"column:user_id;type:uuid;uniqueIndex:idx_user_token,where:deleted_at IS NULL"`
	Token     string          `gorm:"column:token;type:text;uniqueIndex:idx_user_token,where:deleted_at IS NULL"`
	Provider  string          `gorm:"column:provider;type:varchar(255)"`
	LastSeen  time.Time       `gorm:"column:last_seen;autoUpdateTime"`
	CreatedAt time.Time       `gorm:"column:created_at"`
	UpdatedAt time.Time       `gorm:"column:updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"column:deleted_at;"`

	User *User `gorm:"foreignKey:UserID;references:ID"`
}

func (PushToken) TableName() string {
	return "push_tokens"
}

func (s *service) SavePushToken(
	ctx context.Context,
	userID uuid.UUID,
	token string,
	provider usecase.PushProvider) error {
	pt := PushToken{
		UserID:   userID,
		Token:    token,
		Provider: provider.String(),
	}

	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "token"}},
		DoUpdates: clause.AssignmentColumns([]string{"provider", "last_seen", "updated_at"}),
	}).Create(&pt).Error
}

func (s *service) ListPushTokens(ctx context.Context, opt usecase.ListPushTokensOption) ([]usecase.PushToken, int, error) {
	var (
		tokens  []PushToken
		utokens []usecase.PushToken
		count   int64
	)

	db := s.db.Model([]PushToken{}).WithContext(ctx)
	if len(opt.UserIDs) > 0 {
		db = db.Where("user_id IN ?", opt.UserIDs)
	}
	if len(opt.Providers) > 0 {
		db = db.Where("provider IN ?", opt.Providers)
	}

	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	if opt.Limit > 0 {
		db = db.Limit(opt.Limit)
	}
	if opt.Skip > 0 {
		db = db.Offset(opt.Skip)
	}

	if opt.IncludeUser {
		db = db.Preload("User")
	}

	if err := db.Find(&tokens).Error; err != nil {
		return nil, 0, err
	}

	for _, t := range tokens {
		ut := t.ConvertToUsecase()
		if t.User != nil {
			u := t.User.ConvertToUsecase()
			ut.User = &u
		}
		utokens = append(utokens, ut)
	}
	return utokens, int(count), nil
}

func (s *service) DeletePushToken(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Delete(&PushToken{}, "id = ?", id).Error
}

// Convert core model to Usecase
func (u PushToken) ConvertToUsecase() usecase.PushToken {
	var d *time.Time

	if u.DeletedAt != nil {
		d = &u.DeletedAt.Time
	}

	provider, _ := usecase.ParsePushProvider(u.Provider)

	return usecase.PushToken{
		ID:        u.ID,
		UserID:    u.UserID,
		Token:     u.Token,
		Provider:  provider,
		LastSeen:  u.LastSeen,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		DeleteAt:  d,
	}
}
