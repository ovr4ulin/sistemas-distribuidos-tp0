package common

import (
	"errors"
	"strings"
)

type AckMessage struct {
	Success bool
}

func MatchesAckMessage(s string) bool {
	return matchTag(s, "AckMessage")
}

func DeserializeAckMessage(s string) (*AckMessage, error) {
	parts := strings.Split(s, FieldDelimiter)
	if len(parts) != 2 {
		return nil, errors.New("invalid AckMessage format")
	}
	if parts[0] != "AckMessage" {
		return nil, errors.New("invalid tag: expected AckMessage")
	}
	return &AckMessage{Success: parts[1] == "True"}, nil
}

func (a *AckMessage) Serialize() string {
	panic("Not implemented")
}

func (a *AckMessage) Length() int {
	panic("Not implemented")
}
