package firebase

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/librarease/librarease/internal/usecase"

	fb "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/messaging"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/api/option"
)

var path = os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY_PATH")

func New() *Firebase {
	ctx := context.Background()
	sa := option.WithCredentialsFile(path)
	app, err := fb.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}
	message, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
	}

	return &Firebase{auth: client, message: message}
}

type Firebase struct {
	auth    *auth.Client
	message *messaging.Client
}

func (f *Firebase) CreateUser(ctx context.Context, ru usecase.RegisterUser) (string, error) {

	u := &auth.UserToCreate{}
	u.Email(ru.Email)
	u.EmailVerified(false)
	u.Password(ru.Password)
	u.DisplayName(ru.Name)
	u.Disabled(false)

	user, err := f.auth.CreateUser(ctx, u)
	if err != nil {
		return "", err
	}
	fmt.Println("Successfully created user: ", user)

	return user.UID, nil
}

// used by middleware
func (f *Firebase) VerifyIDToken(ctx context.Context, token string) (string, error) {
	t, err := f.auth.VerifyIDToken(ctx, token)
	if err != nil {
		return "", err
	}
	return t.UID, nil
}

func (f *Firebase) SetCustomClaims(ctx context.Context, uid string, claims usecase.CustomClaims) error {
	return f.auth.SetCustomUserClaims(ctx, uid, map[string]any{
		"librarease": map[string]any{
			"id":         claims.ID,
			"role":       claims.Role,
			"admin_libs": claims.AdminLibs,
			"staff_libs": claims.StaffLibs,
		},
	})
}

func (s *Firebase) Provider() usecase.PushProvider {
	return usecase.FCM
}

func (s *Firebase) Send(ctx context.Context, tokens []usecase.PushToken, noti usecase.Notification) error {

	var fcmTokens []string
	for _, t := range tokens {
		if t.Provider == usecase.FCM {
			fcmTokens = append(fcmTokens, t.Token)
		}
	}
	if len(fcmTokens) == 0 {
		return nil
	}

	message := &messaging.MulticastMessage{
		Tokens: fcmTokens,
		Notification: &messaging.Notification{
			Title: noti.Title,
			Body:  noti.Message,
		},
		// Data: noti.Data,
	}

	bres, err := s.message.SendEachForMulticast(ctx, message)
	if err != nil {
		return err
	}
	fmt.Printf("Successfully sent %d messages: %+v\n", bres.SuccessCount, *bres)
	return nil
}
