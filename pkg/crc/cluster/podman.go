package cluster

import (
	"encoding/base64"
	"math/rand"
	"time"
)

func GenerateCockpitBearerToken() string {
	rng := rand.NewSource(time.Now().UnixNano())
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, 8)
	for i := range b {
		b[i] = letterBytes[rng.Int63()%int64(len(letterBytes))] // #nosec
	}
	return base64.StdEncoding.EncodeToString(b)
}
