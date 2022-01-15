package utils

import (
	"os"
	"strings"
	"time"
)

func GetEnv(envName string, valueDefault string) string {
	value := os.Getenv(envName)
	if value == "" {
		return valueDefault
	}
	return value
}

func OnlyAlphaNumberOrUnderscore(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
			return false
		}
	}
	return true
}

func DateToEuropean(t time.Time) string {
	return strings.Replace(t.Format("02-01-2006"), "-", ".", -1)
}
