package push

import (
	"context"

	"github.com/librarease/librarease/internal/usecase"
)

type WebPushSender struct{}

func (s *WebPushSender) Provider() usecase.PushProvider {
	return usecase.WebPush
}

func (s *WebPushSender) Send(ctx context.Context, tokens []usecase.PushToken, noti usecase.Notification) error {
	// TODO: implement WebPush push notification sending
	return nil
}
