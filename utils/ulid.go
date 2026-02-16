package utils

import (
	"crypto/rand"
	"io"
	"time"

	"github.com/oklog/ulid/v2"
)

var entropy io.Reader

func init() {
	entropy = ulid.DefaultEntropy()
}

// GenerateULID ULIDを生成する
func GenerateULID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

// GenerateULIDWithEntropy カスタムエントロピーを使用してULIDを生成する
func GenerateULIDWithEntropy() string {
	customEntropy := rand.Reader
	return ulid.MustNew(ulid.Timestamp(time.Now()), customEntropy).String()
}

// GenerateInviteCode 6桁英数大文字のランダムコードを生成する
func GenerateInviteCode() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b)
}

// CalculateAge 生年月日と現在時刻から年齢を計算する
func CalculateAge(birthDate, now time.Time) int {
	age := now.Year() - birthDate.Year()
	// まだ誕生日が来ていない場合は1歳減らす
	if now.YearDay() < birthDate.YearDay() {
		age--
	}
	return age
}
