package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/trihackathon/api/adapter"
)

func main() {
	// .envファイルを読み込む
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// R2の環境変数を確認
	fmt.Println("=== R2環境変数の確認 ===")
	fmt.Printf("R2_ACCOUNT_ID: %s\n", os.Getenv("R2_ACCOUNT_ID"))
	fmt.Printf("R2_ACCESS_KEY_ID: %s\n", os.Getenv("R2_ACCESS_KEY_ID"))
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	fmt.Printf("R2_SECRET_ACCESS_KEY: %s (長さ: %d)\n",
		strings.Repeat("*", min(len(secretKey), 10)),
		len(secretKey))
	fmt.Printf("R2_BUCKET_NAME: %s\n", os.Getenv("R2_BUCKET_NAME"))
	fmt.Printf("R2_PUBLIC_URL: %s\n", os.Getenv("R2_PUBLIC_URL"))

	// R2に接続テスト
	fmt.Println("\n=== R2接続テスト ===")
	r2 := adapter.NewR2Adapter()

	// テストファイルをアップロード
	testContent := strings.NewReader("test content from API")
	testKey := "test/connection-test.txt"

	fmt.Printf("アップロード中: %s\n", testKey)
	url, err := r2.Upload(context.Background(), testKey, testContent, "text/plain")
	if err != nil {
		log.Fatalf("❌ R2アップロード失敗: %v", err)
	}

	fmt.Printf("✅ R2アップロード成功!\n")
	fmt.Printf("   公開URL: %s\n", url)

	// テストファイルを削除
	fmt.Println("\n削除中...")
	if err := r2.Delete(context.Background(), testKey); err != nil {
		log.Printf("⚠️  R2削除失敗: %v", err)
	} else {
		fmt.Println("✅ R2削除成功")
	}

	fmt.Println("\n=== すべてのテストが完了しました ===")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
