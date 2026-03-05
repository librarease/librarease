package filestorage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	consts "github.com/librarease/librarease/internal/config"
)

type S3FileStorage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
	tempPath      string
	publicPath    string
	region        string
}

func NewS3Storage(bucket string, tempPath string) *S3FileStorage {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}
	client := s3.NewFromConfig(cfg)
	region := cfg.Region
	if region == "" {
		region = "us-east-1"
	}
	return &S3FileStorage{
		client:        client,
		presignClient: s3.NewPresignClient(client),
		bucket:        bucket,
		tempPath:      tempPath,
		publicPath:    "public",
		region:        region,
	}
}

func (f *S3FileStorage) GetTempUploadURL(ctx context.Context, name string) (string, string, error) {
	key := path.Join(f.tempPath, name)
	req, err := f.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: &f.bucket,
		Key:    &key,
	}, func(po *s3.PresignOptions) {
		po.Expires = time.Minute * consts.PRESIGN_URL_EXPIRE_MINUTES
	})
	if err != nil {
		return "", "", err
	}

	return req.URL, key, nil
}

func (f *S3FileStorage) CopyFile(ctx context.Context, source string, dest string) error {
	copySource := f.bucket + "/" + url.PathEscape(source)
	_, err := f.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     &f.bucket,
		CopySource: &copySource,
		Key:        &dest,
	})
	return err
}

func (f *S3FileStorage) CopyFilePreserveFilename(ctx context.Context, source string, dest string) (string, error) {
	filename := source[strings.LastIndex(source, "/")+1:]
	destPath := dest + "/" + filename
	return destPath, f.CopyFile(ctx, source, destPath)
}

func (f *S3FileStorage) MoveTempFilePublic(ctx context.Context, source string, dest string) error {
	return f.MoveTempFile(ctx, source, f.publicPath+"/"+dest)
}

func (f *S3FileStorage) MoveTempFile(ctx context.Context, source string, dest string) error {
	tempSource := f.tempPath + "/" + source
	key := dest + "/" + source
	return f.CopyFile(ctx, tempSource, key)
}

func (f *S3FileStorage) TempPath() string {
	return f.tempPath
}

func (f *S3FileStorage) GetPublicURL(path string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", f.bucket, f.region, path)
}

func (f *S3FileStorage) GetPresignedURL(ctx context.Context, path string) (string, error) {
	req, err := f.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &f.bucket,
		Key:    &path,
	}, func(po *s3.PresignOptions) {
		po.Expires = time.Minute * consts.PRESIGN_URL_EXPIRE_MINUTES
	})
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (f *S3FileStorage) UploadFile(ctx context.Context, path string, data []byte) error {
	contentLength := int64(len(data))
	_, err := f.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        &f.bucket,
		Key:           &path,
		Body:          bytes.NewReader(data),
		ContentLength: aws.Int64(contentLength),
	})
	return err
}

func (f *S3FileStorage) GetReader(ctx context.Context, path string) (io.ReadCloser, error) {
	obj, err := f.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &f.bucket,
		Key:    &path,
	})
	if err != nil {
		return nil, err
	}
	return obj.Body, nil
}
