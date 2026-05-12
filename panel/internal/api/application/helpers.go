package application

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/google/uuid"
)

func genUUID() string { return uuid.New().String() }
func genUUIDShort() string { return uuid.New().String()[:8] }

func randomStr(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, n)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result[i] = chars[num.Int64()]
	}
	return string(result)
}

func strPtr(s string) *string { return &s }
func timePtr(t time.Time) *time.Time { return &t }
