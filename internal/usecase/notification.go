package usecase

import (
	"context"
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

type ListNotificationsOption struct {
	Skip  int
	Limit int
}

func (u Usecase) ListNotifications(ctx context.Context, opt ListNotificationsOption) ([]Notification, int, error) {
	// TODO: implement list notifications
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return nil, 0, fmt.Errorf("user id not found in context")
	}
	fmt.Printf("List notifications for user ID: %s\n", userID.String())
	// notifications, total, err := u.repo.ListNotifications(ctx, opt)
	// if err != nil {
	// 	return nil, 0, err
	// }

	// publicURL, _ := u.fileStorageProvider.GetPublicURL(ctx)

	// var list []Notification
	// for _, n := range notifications {
	// 	if n.ReferenceID != nil {
	// 		n.ReferenceID = &uuid.UUID{}
	// 	}
	// 	list = append(list, n)
	// }

	// return list, total, err

	return nil, 0, nil
}

func (u Usecase) ReadNotification(ctx context.Context, id uuid.UUID) error {
	// TODO: implement read notification
	fmt.Printf("Read notification with ID: %s\n", id.String())
	// return u.repo.ReadNotification(ctx, id)
	return nil
}

func (u Usecase) ReadAllNotifications(ctx context.Context) error {
	// TODO: implement read all notifications
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return fmt.Errorf("user id not found in context")
	}
	fmt.Printf("Read all notifications for user ID: %s\n", userID.String())
	// return u.repo.ReadAllNotifications(ctx, userID)
	return nil
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
		defer close(inbound)

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
