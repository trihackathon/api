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
	ctx := context.Background()

	var app *firebase.App
	var err error

	// Cloud RunのGCPプロジェクトとFirebaseプロジェクトが異なる場合に対応
	// FIREBASE_PROJECT_ID で明示的に指定可能
	var config *firebase.Config
	if projectID := os.Getenv("FIREBASE_PROJECT_ID"); projectID != "" {
		config = &firebase.Config{ProjectID: projectID}
		log.Printf("Firebase: using explicit project ID: %s", projectID)
	}

	// GOOGLE_APPLICATION_CREDENTIALS が設定されていればファイルを使う（ローカル開発用）
	// 未設定の場合はGCPのApplication Default Credentialsを使う（Cloud Run等）
	if credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); credPath != "" {
		opt := option.WithCredentialsFile(credPath)
		app, err = firebase.NewApp(ctx, config, opt)
	} else {
		app, err = firebase.NewApp(ctx, config)
	}
	if err != nil {
		log.Fatalf("Firebase初期化エラー: %v", err)
	}

	authClient, err := app.Auth(ctx)
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
