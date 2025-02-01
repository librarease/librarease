package filestorage

import (
	"context"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	consts "github.com/librarease/librarease/internal/config"
)

type FileStorage struct {
	client   *s3.Client
	bucket   string
	tempPath string
}

func New(bucket string, tempPath string) *FileStorage {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}
	return &FileStorage{
		client:   s3.NewFromConfig(cfg),
		bucket:   bucket,
		tempPath: tempPath,
	}
}

func (f *FileStorage) GetTempUploadURL(ctx context.Context, name string) (string, error) {
	key := path.Join(f.tempPath, name)
	presignClient := s3.NewPresignClient(f.client)

	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: &f.bucket,
		Key:    &key,
	}, func(po *s3.PresignOptions) {
		po.Expires = time.Minute * consts.PRESIGN_URL_EXPIRE_MINUTES
	})
	if err != nil {
		return "", err
	}

	return req.URL, nil
}
