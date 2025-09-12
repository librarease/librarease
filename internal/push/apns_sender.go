package push

import (
	"context"

	"github.com/librarease/librarease/internal/usecase"
)

type APNSSender struct{}

func (s *APNSSender) Provider() usecase.PushProvider {
	return usecase.APNS
}

func (s *APNSSender) Send(ctx context.Context, tokens []usecase.PushToken, noti usecase.Notification) error {
	// TODO: implement APNS push notification sending
	return nil
}
