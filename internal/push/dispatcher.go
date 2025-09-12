package push

import (
	"context"

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
	for _, sender := range d.senders {
		if err := sender.Send(ctx, tokens, noti); err != nil {
			return err
		}
	}
	return nil
}

type PushSender interface {
	Provider() usecase.PushProvider
	Send(context.Context, []usecase.PushToken, usecase.Notification) error
}
