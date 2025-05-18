package database

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/librarease/librarease/internal/usecase"
	"gorm.io/gorm"
)

type Notification struct {
	ID            uuid.UUID  `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID        uuid.UUID  `gorm:"column:user_id;type:uuid;" json:"user_id"`
	User          *User      `gorm:"foreignKey:UserID;references:ID"`
	Title         string     `gorm:"column:title" json:"title"`
	Message       string     `gorm:"column:message" json:"message"`
	CreatedAt     time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at" json:"updated_at"`
	ReadAt        *time.Time `gorm:"column:read_at" json:"read_at"`
	ReferenceID   *uuid.UUID `gorm:"column:reference_id;type:uuid" json:"reference_id"`
	ReferenceType string     `gorm:"column:reference_type" json:"reference_type"`
	DeletedAt     *gorm.DeletedAt
}

func (Notification) TableName() string {
	return "notifications"
}

// Convert core model to Usecase
func (n Notification) ConvertToUsecase() usecase.Notification {
	var d *time.Time
	if n.DeletedAt != nil {
		d = &n.DeletedAt.Time
	}
	return usecase.Notification{
		ID:            n.ID,
		UserID:        n.UserID,
		Title:         n.Title,
		Message:       n.Message,
		CreatedAt:     n.CreatedAt,
		UpdatedAt:     n.UpdatedAt,
		ReadAt:        n.ReadAt,
		ReferenceID:   n.ReferenceID,
		ReferenceType: n.ReferenceType,
		DeletedAt:     d,
	}
}

type notificationHub struct {
	mu          sync.Mutex
	subscribers map[chan<- usecase.Notification]struct{}
	conn        *pgx.Conn
}

func NewNotificationHub(conn *pgx.Conn) *notificationHub {
	hub := &notificationHub{
		conn:        conn,
		subscribers: make(map[chan<- usecase.Notification]struct{}),
	}
	go hub.listen()
	return hub
}

func (h *notificationHub) listen() {
	ctx := context.Background()
	for {
		n, err := h.conn.WaitForNotification(ctx)
		if err != nil {
			fmt.Printf("Error waiting for notification: %v\n", err)
			return
		}

		if n == nil {
			continue
		}
		// Parse n.Payload into usecase.Notification (implement your own parsing)
		notif := parseNotification(n)

		// Notify all subscribers
		h.mu.Lock()
		for ch := range h.subscribers {
			select {
			case ch <- notif:
			default:
				// If the channel is full, we skip sending the notification
				// to avoid blocking the notification hub.
				// You might want to log this or handle it differently.
				fmt.Printf("Subscriber channel is full, skipping notification: %v\n", notif)
			}
		}
		h.mu.Unlock()
	}
}

func (h *notificationHub) Subscribe(ch chan<- usecase.Notification) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.subscribers[ch] = struct{}{}
}

func (h *notificationHub) Unsubscribe(ch chan<- usecase.Notification) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.subscribers, ch)
}

// Helper to parse pgx notification payload to usecase.Notification
func parseNotification(n *pgconn.Notification) usecase.Notification {
	var notification Notification
	if err := json.Unmarshal([]byte(n.Payload), &notification); err != nil {
		fmt.Printf("Error parsing notification: %v\n", err)
		return usecase.Notification{}
	}

	return notification.ConvertToUsecase()
}

// Implement the repo interface:
func (s *service) SubscribeNotifications(ctx context.Context, ch chan<- usecase.Notification) error {
	s.noti.Subscribe(ch)
	return nil
}

func (s *service) UnsubscribeNotifications(ctx context.Context, ch chan<- usecase.Notification) error {
	s.noti.Unsubscribe(ch)
	return nil
}

func (s *service) ListNotifications(ctx context.Context, opt usecase.ListNotificationsOption) ([]usecase.Notification, int, int, error) {

	var (
		notifications []Notification
		total         int64
	)

	query := s.db.WithContext(ctx).Model(&Notification{}).
		Where("user_id = ?", opt.UserID).
		Order("created_at desc").
		Limit(opt.Limit).
		Offset(opt.Skip)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}

	if err := query.Find(&notifications).Error; err != nil {
		return nil, 0, 0, err
	}

	var unreadCount int64
	if err := s.db.WithContext(ctx).Model(&Notification{}).
		Where("user_id = ? AND read_at IS NULL", opt.UserID).
		Count(&unreadCount).Error; err != nil {
		return nil, 0, 0, err
	}

	result := make([]usecase.Notification, len(notifications))
	for i, n := range notifications {
		result[i] = n.ConvertToUsecase()
	}

	return result, int(unreadCount), int(total), nil
}

func (s *service) ReadNotification(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).
		Model(&Notification{}).
		Where("id = ?", id).
		Update("read_at", time.Now()).Error
}

func (s *service) ReadAllNotifications(ctx context.Context, userID uuid.UUID) error {
	return s.db.WithContext(ctx).
		Model(&Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", time.Now()).Error
}

func (s *service) CountUnreadNotifications(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64
	if err := s.db.WithContext(ctx).
		Model(&Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}
