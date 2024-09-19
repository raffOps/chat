package pkg

import (
	"github.com/google/uuid"
	"github.com/raffops/chat/internal/app/chat"
)

type uuidGenerator struct {
}

func (u uuidGenerator) GenerateUuid(namespace uuid.UUID, data []byte) uuid.UUID {
	return uuid.NewSHA1(namespace, data)
}

func NewUuidGenerator() chat.UuidGenerator {
	return &uuidGenerator{}
}
