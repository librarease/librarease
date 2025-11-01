package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/librarease/librarease/internal/config"
)

func (u Usecase) GetTempUploadURL(ctx context.Context, name string) (string, string, error) {
	userID, ok := ctx.Value(config.CTX_KEY_USER_ID).(uuid.UUID)
	if !ok {
		return "", "", fmt.Errorf("user id not found in context")
	}
	path := fmt.Sprintf("%s-%d/%s", userID.String()[:8], time.Now().Unix(), name)
	return u.fileStorageProvider.GetTempUploadURL(ctx, path)
}
