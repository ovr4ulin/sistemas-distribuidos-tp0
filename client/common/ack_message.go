package common

import (
	"errors"
	"fmt"
	"strings"
)

type AckMessage struct {
	ProcessedCount string
}

func (a *AckMessage) Serialize() string {
	return fmt.Sprintf("AckMessage^%s", a.ProcessedCount)
}

func MatchesAckMessage(s string) bool {
	return matchTag(s, "AckMessage")
}

func DeserializeAckMessage(s string) (*AckMessage, error) {
	parts := strings.Split(s, "^")
	if len(parts) != 2 {
		return nil, errors.New("invalid AckMessage format")
	}
	if parts[0] != "AckMessage" {
		return nil, errors.New("invalid tag: expected AckMessage")
	}
	return &AckMessage{ProcessedCount: parts[1]}, nil
}
