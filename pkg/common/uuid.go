package common

import "github.com/google/uuid"

func GenUUid() string {
	return uuid.New().String()
}
