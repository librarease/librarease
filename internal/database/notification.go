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
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type NotificationEvent struct {
	ID            uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()" json:"event_id"`
	Title         string          `gorm:"column:title" json:"title"`
	Message       string          `gorm:"column:message" json:"message"`
	ReferenceID   *uuid.UUID      `gorm:"column:reference_id;type:uuid" json:"reference_id"`
	ReferenceType string          `gorm:"column:reference_type" json:"reference_type"`
	GroupKey      string          `gorm:"column:group_key" json:"group_key"`
	Metadata      datatypes.JSON  `gorm:"column:metadata;type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt     time.Time       `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time       `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt     *gorm.DeletedAt `json:"deleted_at"`
}

func (NotificationEvent) TableName() string {
	return "notification_events"
}

type NotificationRecipient struct {
	ID        uuid.UUID          `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	EventID   uuid.UUID          `gorm:"column:event_id;type:uuid;uniqueIndex:idx_notification_recipients_event_user" json:"event_id"`
	Event     *NotificationEvent `gorm:"foreignKey:EventID;references:ID" json:"event"`
	UserID    uuid.UUID          `gorm:"column:user_id;type:uuid;uniqueIndex:idx_notification_recipients_event_user" json:"user_id"`
	User      *User              `gorm:"foreignKey:UserID;references:ID" json:"user"`
	ReadAt    *time.Time         `gorm:"column:read_at" json:"read_at"`
	SentAt    *time.Time         `gorm:"column:sent_at" json:"sent_at"`
	CreatedAt time.Time          `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time          `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *gorm.DeletedAt    `json:"deleted_at"`
}

func (NotificationRecipient) TableName() string {
	return "notification_recipients"
}

func (r NotificationRecipient) ConvertToUsecase() usecase.Notification {
	n := usecase.Notification{
		ID:        r.ID,
		EventID:   r.EventID,
		UserID:    r.UserID,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
		ReadAt:    r.ReadAt,
	}
	if r.DeletedAt != nil {
		n.DeletedAt = &r.DeletedAt.Time
	}
	if r.Event != nil {
		n.Title = r.Event.Title
		n.Message = r.Event.Message
		n.ReferenceID = r.Event.ReferenceID
		n.ReferenceType = r.Event.ReferenceType
		if r.Event.CreatedAt.Before(n.CreatedAt) {
			n.CreatedAt = r.Event.CreatedAt
		}
		if r.Event.UpdatedAt.After(n.UpdatedAt) {
			n.UpdatedAt = r.Event.UpdatedAt
		}
	}
	return n
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

	defer func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		for subscriber := range h.subscribers {
			close(subscriber)
			delete(h.subscribers, subscriber)
		}
	}()

	for {
		n, err := h.conn.WaitForNotification(ctx)
		if err != nil {
			fmt.Printf("Error waiting for notification: %v\n", err)
			return
		}

		if n == nil {
			continue
		}
		notif := parseNotification(n)

		h.mu.Lock()
		for ch := range h.subscribers {
			select {
			case ch <- notif:
			default:
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

func parseNotification(n *pgconn.Notification) usecase.Notification {
	var notification usecase.Notification
	if err := json.Unmarshal([]byte(n.Payload), &notification); err != nil {
		fmt.Printf("Error parsing notification: %v\n", err)
		return usecase.Notification{}
	}

	return notification
}

func (s *service) SubscribeNotifications(ctx context.Context, ch chan<- usecase.Notification) error {
	s.noti.Subscribe(ch)
	return nil
}

func (s *service) UnsubscribeNotifications(ctx context.Context, ch chan<- usecase.Notification) error {
	s.noti.Unsubscribe(ch)
	return nil
}

func (s *service) GetNotification(ctx context.Context, id uuid.UUID) (usecase.Notification, error) {
	var recipient NotificationRecipient
	if err := s.db.WithContext(ctx).
		Preload("Event").
		Where("notification_recipients.id = ? OR notification_recipients.event_id = ?", id, id).
		Order("notification_recipients.created_at ASC").
		First(&recipient).
		Error; err != nil {
		return usecase.Notification{}, err
	}
	return recipient.ConvertToUsecase(), nil
}

func (s *service) ListNotificationRecipients(ctx context.Context, eventID uuid.UUID) ([]usecase.Notification, error) {
	var recipients []NotificationRecipient
	if err := s.db.WithContext(ctx).
		Preload("Event").
		Where("event_id = ?", eventID).
		Find(&recipients).
		Error; err != nil {
		return nil, err
	}

	result := make([]usecase.Notification, len(recipients))
	for i, r := range recipients {
		result[i] = r.ConvertToUsecase()
	}
	return result, nil
}

func (s *service) ListNotifications(ctx context.Context, opt usecase.ListNotificationsOption) ([]usecase.Notification, int, int, error) {
	var (
		recipients []NotificationRecipient
		total      int64
	)

	base := s.db.
		WithContext(ctx).
		Model(&NotificationRecipient{}).
		Where("user_id = ?", opt.UserID)

	if opt.IsUnread {
		base = base.Where("read_at IS NULL")
	}

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}

	query := base.
		Preload("Event").
		Order("created_at desc")

	if opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	}

	if opt.Skip > 0 {
		query = query.Offset(opt.Skip)
	}

	if err := query.Find(&recipients).Error; err != nil {
		return nil, 0, 0, err
	}

	var unreadCount int64
	if err := s.db.WithContext(ctx).Model(&NotificationRecipient{}).
		Where("user_id = ? AND read_at IS NULL", opt.UserID).
		Count(&unreadCount).Error; err != nil {
		return nil, 0, 0, err
	}

	result := make([]usecase.Notification, len(recipients))
	for i, r := range recipients {
		result[i] = r.ConvertToUsecase()
	}

	return result, int(unreadCount), int(total), nil
}

func (s *service) ReadNotification(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).
		Model(&NotificationRecipient{}).
		Where("id = ?", id).
		Update("read_at", time.Now()).Error
}

func (s *service) ReadAllNotifications(ctx context.Context, userID uuid.UUID) error {
	return s.db.WithContext(ctx).
		Model(&NotificationRecipient{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", time.Now()).Error
}

func (s *service) CountUnreadNotifications(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64
	if err := s.db.WithContext(ctx).
		Model(&NotificationRecipient{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (s *service) CreateNotification(ctx context.Context, n usecase.Notification) (usecase.Notification, error) {
	recipients := n.RecipientIDs
	if len(recipients) == 0 && n.UserID != uuid.Nil {
		recipients = uuid.UUIDs{n.UserID}
	}
	seen := make(map[uuid.UUID]struct{}, len(recipients))
	uniqueRecipients := make(uuid.UUIDs, 0, len(recipients))
	for _, userID := range recipients {
		if userID == uuid.Nil {
			continue
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		uniqueRecipients = append(uniqueRecipients, userID)
	}
	recipients = uniqueRecipients
	if len(recipients) == 0 {
		return usecase.Notification{}, fmt.Errorf("notification must have at least one recipient")
	}

	var created NotificationRecipient
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		event := NotificationEvent{
			Title:         n.Title,
			Message:       n.Message,
			ReferenceID:   n.ReferenceID,
			ReferenceType: n.ReferenceType,
		}
		if err := tx.Clauses(clause.Returning{}).Create(&event).Error; err != nil {
			return err
		}

		rows := make([]NotificationRecipient, 0, len(recipients))
		for _, userID := range recipients {
			rows = append(rows, NotificationRecipient{
				EventID: event.ID,
				UserID:  userID,
				ReadAt:  n.ReadAt,
				Event:   &event,
			})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return err
		}
		created = rows[0]
		created.Event = &event
		return nil
	})
	if err != nil {
		return usecase.Notification{}, err
	}

	return created.ConvertToUsecase(), nil
}
