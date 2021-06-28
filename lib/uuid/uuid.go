package uuid

import (
	"crypto/rand"

	"github.com/google/uuid"
	"github.com/oklog/ulid"
)

func UUID() string {
	return uuid.New().String()
}

func UUID8() string {
	return UUID()[:8]
}

func ULID() (string, error) {
	id, err := ulid.New(ulid.Now(), rand.Reader)
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
