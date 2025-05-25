package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/librarease/librarease/internal/usecase"
)

type Notification struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Message       string  `json:"message"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	ReadAt        *string `json:"read_at,omitempty"`
	ReferenceID   *string `json:"reference_id,omitempty"`
	ReferenceType string  `json:"reference_type"`
}

type ListNotificationRequest struct {
	Skip  int `query:"skip"`
	Limit int `query:"limit" validate:"required,min=1,max=100"`
}

func (s *Server) ListNotifications(ctx echo.Context) error {
	var req ListNotificationRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}

	notifications, unread, total, err := s.server.ListNotifications(ctx.Request().Context(), usecase.ListNotificationsOption{
		Skip:  req.Skip,
		Limit: req.Limit,
	})
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	list := make([]Notification, 0, len(notifications))

	for _, n := range notifications {
		var (
			readAt *string
			refID  *string
		)
		if n.ReadAt != nil {
			t := n.ReadAt.Format(time.RFC3339)
			readAt = &t
		}
		if n.ReferenceID != nil {
			id := n.ReferenceID.String()
			refID = &id
		}
		list = append(list, Notification{
			ID:            n.ID.String(),
			Title:         n.Title,
			Message:       n.Message,
			CreatedAt:     n.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     n.UpdatedAt.Format(time.RFC3339),
			ReadAt:        readAt,
			ReferenceID:   refID,
			ReferenceType: n.ReferenceType,
		})
	}

	return ctx.JSON(200, Res{
		Data: list,
		Meta: &Meta{
			Unread: &unread,
			Total:  total,
			Skip:   req.Skip,
			Limit:  req.Limit,
		},
	})
}

type ReadNotificationRequest struct {
	ID string `param:"id" validate:"required,uuid"`
}

func (s *Server) ReadNotification(ctx echo.Context) error {
	var req ReadNotificationRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}
	id, _ := uuid.Parse(req.ID)

	err := s.server.ReadNotification(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.NoContent(204)
}

func (s *Server) ReadAllNotifications(ctx echo.Context) error {
	err := s.server.ReadAllNotifications(ctx.Request().Context())
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	return ctx.NoContent(204)
}

type StreamNotificationsRequest struct {
	UserID string `query:"user_id" validate:"required,uuid"`
}

// REF: https://echo.labstack.com/docs/cookbook/sse
func (s *Server) StreamNotifications(ctx echo.Context) error {
	var req StreamNotificationsRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(400, map[string]string{"error": err.Error()})
	}
	if err := s.validator.Struct(req); err != nil {
		return ctx.JSON(422, map[string]string{"error": err.Error()})
	}
	userID, _ := uuid.Parse(req.UserID)
	ch, err := s.server.StreamNotifications(ctx.Request().Context(), userID)
	if err != nil {
		return ctx.JSON(500, map[string]string{"error": err.Error()})
	}

	w := ctx.Response()
	w.Header().Set(echo.HeaderContentType, "text/event-stream")
	w.Header().Set(echo.HeaderCacheControl, "no-cache, no-store, no-transform")
	w.Header().Set(echo.HeaderConnection, "keep-alive")

	var noti *Notification

	// to prevent pending when no notification on connection
	// w.Write([]byte("\n\n"))
	// w.Flush()

	for {
		select {
		case <-ctx.Request().Context().Done():
			return nil
		case msg, ok := <-ch:
			if !ok {
				fmt.Printf("[DEBUG] notification stream closed\n")
				return nil
			}
			fmt.Printf("[DEBUG] received notification: %v\n", msg)
			if msg.ID == uuid.Nil {
				continue
			}

			var referenceID *string
			if msg.ReferenceID != nil {
				id := msg.ReferenceID.String()
				referenceID = &id
			}

			noti = &Notification{
				ID:            msg.ID.String(),
				Title:         msg.Title,
				Message:       msg.Message,
				CreatedAt:     msg.CreatedAt.Format(time.RFC3339),
				UpdatedAt:     msg.UpdatedAt.Format(time.RFC3339),
				ReferenceID:   referenceID,
				ReferenceType: msg.ReferenceType,
			}
			if msg.ReadAt != nil {
				t := msg.ReadAt.Format(time.RFC3339)
				noti.ReadAt = &t
			}
			if msg.ReferenceID != nil {
				id := msg.ReferenceID.String()
				noti.ReferenceID = &id
			}
			data, err := json.Marshal(noti)
			noti = nil

			if err != nil {
				fmt.Printf("error marshalling notification: %v\n", err)
				continue
			}

			w.Write([]byte("data: " + string(data) + "\n\n"))
			w.Flush()
		}
	}
}
