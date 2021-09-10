package util

import (
	"math/rand"
	"os"
	"strconv"
	"time"
)

func EnsureBaseDirectoriesExist(path string) error {
	return os.MkdirAll(path, 0750)
}

func GetHomeDir() string {
	return os.Getenv("HOME")
}

func GenerateCorrelation() string {
	rand.Seed(time.Now().UnixNano())
	return strconv.Itoa(1000000 + rand.Intn(8000000))
}
