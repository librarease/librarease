package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/librarease/librarease/internal/config"
)

type PushProvider uint

const (
	_ PushProvider = iota
	FCM
	APNS
	WebPush
)

var pushProviderMap = map[PushProvider]string{
	FCM:     "fcm",
	APNS:    "apns",
	WebPush: "webpush",
}

var pushProviderReverseMap = map[string]PushProvider{
	"fcm":     FCM,
	"apns":    APNS,
	"webpush": WebPush,
}

func ParsePushProvider(s string) (PushProvider, bool) {
	p, ok := pushProviderReverseMap[s]
	return p, ok
}

func (p PushProvider) String() string {
	return pushProviderMap[p]
}

type PushToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	Provider  PushProvider
	LastSeen  time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	DeleteAt  *time.Time

	User *User
}

type ListPushTokensOption struct {
	Skip        int
	Limit       int
	UserIDs     uuid.UUIDs
	Providers   []string
	IncludeUser bool
}

func (u Usecase) SavePushToken(ctx context.Context, token string, provider PushProvider) error {
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return fmt.Errorf("user id not found in context")
	}

	return u.repo.SavePushToken(ctx, userID, token, provider)
}
