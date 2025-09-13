package push

import (
	"context"
	"maps"

	"github.com/google/uuid"
	"github.com/librarease/librarease/internal/usecase"
)

func NewPushDispatcher(senders ...PushSender) *PushDispatcher {
	m := make(map[usecase.PushProvider]PushSender)
	for _, s := range senders {
		m[s.Provider()] = s
	}
	return &PushDispatcher{senders: m}
}

type PushDispatcher struct {
	senders map[usecase.PushProvider]PushSender
}

func (d *PushDispatcher) Send(ctx context.Context, tokens []usecase.PushToken, noti usecase.Notification) error {
	mergedInvalids := make(map[uuid.UUID]string)
	for _, sender := range d.senders {
		if err := sender.Send(ctx, tokens, noti); err != nil {
			if inv, ok := err.(usecase.InvalidTokenError); ok {
				maps.Copy(mergedInvalids, inv)
				continue
			}
			return err
		}
	}
	if len(mergedInvalids) > 0 {
		return usecase.NewInvalidTokenError(mergedInvalids)
	}
	return nil
}

type PushSender interface {
	Provider() usecase.PushProvider
	Send(context.Context, []usecase.PushToken, usecase.Notification) error
}
