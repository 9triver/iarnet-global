package util

import (
	"github.com/lithammer/shortuuid/v4"
)

// GenID 生成唯一 ID
func GenID() string {
	return shortuuid.New()
}

// GenIDWith 生成带前缀的唯一 ID
func GenIDWith(prefix string) string {
	return prefix + shortuuid.New()
}
