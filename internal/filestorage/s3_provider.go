package filestorage

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	consts "github.com/librarease/librarease/internal/config"
)

type S3FileStorage struct {
	client   *s3.Client
	bucket   string
	tempPath string
}

func NewS3Storage(bucket string, tempPath string) *S3FileStorage {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}
	return &S3FileStorage{
		client:   s3.NewFromConfig(cfg),
		bucket:   bucket,
		tempPath: tempPath,
	}
}

func (f *S3FileStorage) GetTempUploadURL(ctx context.Context, name string) (string, error) {
	var (
		key           = path.Join(f.tempPath, name)
		presignClient = s3.NewPresignClient(f.client)
	)
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

func (f *S3FileStorage) MoveTempFile(ctx context.Context, source string, dest string) error {
	var (
		tempSource = f.bucket + "/" + f.tempPath + "/" + source
		key        = dest + "/" + source
	)
	_, err := f.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     &f.bucket,
		CopySource: &tempSource,
		Key:        &key,
	})
	return err
}

func (f *S3FileStorage) GetPublicURL(_ context.Context) (string, error) {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", f.bucket, "ap-southeast-1", "public"), nil
}
