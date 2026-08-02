package generate

import "github.com/google/uuid"

func ID() uuid.UUID {
	uuidV7, _ := uuid.NewV7()
	return uuidV7
}
