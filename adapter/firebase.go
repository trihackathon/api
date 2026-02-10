package adapter

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type FirebaseAdapter struct {
	AuthClient *auth.Client
}

func NewFirebaseAdapter() *FirebaseAdapter {
	credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credPath == "" {
		log.Fatal("GOOGLE_APPLICATION_CREDENTIALS environment variable is not set")
	}

	opt := option.WithCredentialsFile(credPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Firebase初期化エラー: %v", err)
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("Firebase Auth初期化エラー: %v", err)
	}

	return &FirebaseAdapter{AuthClient: authClient}
}

// VerifyToken はIDトークンを検証し、トークン情報を返す
func (f *FirebaseAdapter) VerifyToken(ctx context.Context, idToken string) (*auth.Token, error) {
	token, err := f.AuthClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// CreateCustomToken はテスト用のカスタムトークンを生成する
func (f *FirebaseAdapter) CreateCustomToken(ctx context.Context, uid string) (string, error) {
	return f.AuthClient.CustomToken(ctx, uid)
}
