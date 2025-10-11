package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/librarease/librarease/internal/config"
)

type Notification struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Title         string
	Message       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ReadAt        *time.Time
	ReferenceID   *uuid.UUID
	ReferenceType string
	DeletedAt     *time.Time
}

type NotificationDispatcher interface {
	Send(context.Context, []PushToken, Notification) error
}

type ListNotificationsOption struct {
	Skip     int
	Limit    int
	UserID   uuid.UUID
	IsUnread bool
}

func (u Usecase) ListNotifications(ctx context.Context, opt ListNotificationsOption) ([]Notification, int, int, error) {
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return nil, 0, 0, fmt.Errorf("user id not found in context")
	}
	return u.repo.ListNotifications(ctx, ListNotificationsOption{
		Skip:     opt.Skip,
		Limit:    opt.Limit,
		UserID:   userID,
		IsUnread: opt.IsUnread,
	})
}

func (u Usecase) ReadNotification(ctx context.Context, id uuid.UUID) error {
	return u.repo.ReadNotification(ctx, id)
}

func (u Usecase) ReadAllNotifications(ctx context.Context) error {
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return fmt.Errorf("user id not found in context")
	}
	return u.repo.ReadAllNotifications(ctx, userID)
}

// REF:
// https://github.dev/brojonat/notifier
// https://brandur.org/notifier
// https://www.finly.ch/engineering-blog/436253-building-a-real-time-notification-system-in-go-with-postgresql

// StreamNotifications creates a notification stream for the specified user.
// It filters notifications based on the userID and handles cleanup when the context is done.
func (u Usecase) StreamNotifications(ctx context.Context, userID uuid.UUID) (<-chan Notification, error) {
	inbound := make(chan Notification, 10)
	if err := u.repo.SubscribeNotifications(ctx, inbound); err != nil {
		close(inbound)
		return nil, fmt.Errorf("subscribe to notifications: %w", err)
	}

	notifications := make(chan Notification, 10)
	go func() {
		defer close(notifications)
		defer u.repo.UnsubscribeNotifications(ctx, inbound)
		// NOTE: inbound will be closed by notificationHub
		// defer close(inbound)

		for {
			select {
			case <-ctx.Done():
				return
			case n, ok := <-inbound:
				if !ok {
					return
				}
				// Only forward notifications for the specified user
				if n.UserID == userID {
					// Non-blocking send to avoid slow consumers
					select {
					case notifications <- n:
					default:
						fmt.Printf("dropping notification for user %s: channel full\n", userID)
					}
				}
			}
		}
	}()

	return notifications, nil
}

type InvalidTokenError map[uuid.UUID]string

func (e InvalidTokenError) Error() string {
	var msg string
	for k, v := range e {
		msg += fmt.Sprintf("TokenID: %s, Reason: %s; ", k, v)
	}

	return fmt.Sprintf("invalid tokens: %s", msg)
}

func NewInvalidTokenError(m map[uuid.UUID]string) InvalidTokenError {
	return InvalidTokenError(m)
}

func (u Usecase) CreateNotification(ctx context.Context, n Notification) error {
	noti, err := u.repo.CreateNotification(ctx, n)
	if err != nil {
		return err
	}

	tokens, _, err := u.repo.ListPushTokens(ctx, ListPushTokensOption{
		UserIDs: uuid.UUIDs{n.UserID},
	})
	if err != nil {
		return err
	}

	if err := u.dispatcher.Send(ctx, tokens, noti); err != nil {
		var invalidErr InvalidTokenError
		if errors.As(err, &invalidErr) {
			for id, reason := range invalidErr {
				fmt.Printf("deleting invalid token %s: %s\n", id, reason)
				if err := u.repo.DeletePushToken(ctx, id); err != nil {
					fmt.Printf("failed to delete invalid token %s: %v\n", id, err)
				}
			}
			// return nil because the notification is created successfully
			return nil
		}
		return fmt.Errorf("send notification: %w", err)
	}
	return nil
}
