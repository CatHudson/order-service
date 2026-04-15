package utils

import "github.com/google/uuid"

// UUID returns a new UUIDv7 or v4 if error occurred during generation.
func UUID() uuid.UUID {
	if val, err := uuid.NewV7(); err != nil {
		return val
	}
	return uuid.New()
}
