package filestorage

import (
	"context"
	"fmt"
	"time"

	consts "github.com/librarease/librarease/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinIOStorage(bucket, tempPath, publicPath, endpoint, accessKeyID, secretAccessKey string) *MinIOStorage {
	m, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		panic(err)
	}
	return &MinIOStorage{
		client:     m,
		bucket:     bucket,
		tempPath:   tempPath,
		publicPath: publicPath,
	}
}

type MinIOStorage struct {
	client     *minio.Client
	bucket     string
	tempPath   string
	publicPath string
}

func (f *MinIOStorage) GetPublicURL(_ context.Context) (string, error) {
	return fmt.Sprintf("%s/%s/%s", f.client.EndpointURL(), f.bucket, f.publicPath), nil
}

func (f *MinIOStorage) GetTempUploadURL(ctx context.Context, name string) (string, error) {
	u, err := f.client.PresignedPutObject(ctx, f.bucket, f.tempPath+"/"+name, time.Minute*consts.PRESIGN_URL_EXPIRE_MINUTES)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (f *MinIOStorage) MoveTempFilePublic(ctx context.Context, source string, dest string) error {
	var (
		tempSource = f.tempPath + "/" + source
		key        = f.publicPath + "/" + dest + "/" + source
	)
	copyDest := minio.CopyDestOptions{
		Bucket: f.bucket,
		Object: key,
	}
	copySource := minio.CopySrcOptions{
		Bucket: f.bucket,
		Object: tempSource,
	}
	_, err := f.client.CopyObject(ctx, copyDest, copySource)
	return err
}

func (f *MinIOStorage) TempPath() string {
	return f.tempPath
}

func (f *MinIOStorage) GetPresignedURL(ctx context.Context, path string) (string, error) {
	u, err := f.client.PresignedGetObject(ctx, f.bucket, path, time.Minute*consts.PRESIGN_URL_EXPIRE_MINUTES, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
