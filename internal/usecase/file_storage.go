package usecase

import "context"

func (u Usecase) GetTempUploadURL(ctx context.Context, name string) (string, error) {
	return u.fileStorageProvider.GetTempUploadURL(ctx, name)
}
