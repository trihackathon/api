package adapter

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Adapter struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

func NewR2Adapter() *R2Adapter {
	accountID := os.Getenv("R2_ACCOUNT_ID")
	accessKeyID := os.Getenv("R2_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucket := os.Getenv("R2_BUCKET_NAME")
	publicURL := os.Getenv("R2_PUBLIC_URL")

	if accountID == "" || accessKeyID == "" || secretAccessKey == "" || bucket == "" {
		log.Fatal("R2環境変数が設定されていません: R2_ACCOUNT_ID, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, R2_BUCKET_NAME")
	}

	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	client := s3.New(s3.Options{
		Region:      "auto",
		Credentials: credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		BaseEndpoint: &endpoint,
	})

	return &R2Adapter{
		client:    client,
		bucket:    bucket,
		publicURL: publicURL,
	}
}

// Upload はファイルをR2にアップロードし、公開URLを返す
func (r *R2Adapter) Upload(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &r.bucket,
		Key:         &key,
		Body:        body,
		ContentType: &contentType,
	})
	if err != nil {
		return "", fmt.Errorf("R2アップロードエラー: %w", err)
	}

	url := fmt.Sprintf("%s/%s", r.publicURL, key)
	return url, nil
}

// Delete はR2からファイルを削除する
func (r *R2Adapter) Delete(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &r.bucket,
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("R2削除エラー: %w", err)
	}
	return nil
}
