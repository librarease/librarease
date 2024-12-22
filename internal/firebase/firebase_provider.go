package firebase

import (
	"context"
	"fmt"
	"librarease/internal/usecase"
	"log"
	"os"

	fb "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
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

	return &Firebase{client}
}

type Firebase struct {
	client *auth.Client
}

func (f *Firebase) CreateUser(ctx context.Context, ru usecase.RegisterUser) (string, error) {

	u := &auth.UserToCreate{}
	u.Email(ru.Email)
	u.EmailVerified(false)
	u.Password(ru.Password)
	u.DisplayName(ru.Name)
	u.Disabled(false)

	user, err := f.client.CreateUser(ctx, u)
	if err != nil {
		return "", err
	}
	fmt.Println("Successfully created user: ", user)

	return user.UID, nil
}

// used by middleware
func (f *Firebase) VerifyIDToken(ctx context.Context, token string) (string, error) {
	t, err := f.client.VerifyIDToken(ctx, token)
	if err != nil {
		return "", err
	}
	return t.UID, nil
}
