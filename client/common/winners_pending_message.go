package common

import (
	"errors"
	"strings"
)

type WinnersPendingMessage struct {
}

func MatchesWinnersPendingMessage(s string) bool {
	return matchTag(s, "WinnersPendingMessage")
}

func DeserializeWinnersPendingMessage(s string) (*WinnersPendingMessage, error) {
	parts := strings.Split(s, FieldDelimiter)
	if len(parts) != 1 {
		return nil, errors.New("invalid WinnersPendingMessage format")
	}
	if parts[0] != "WinnersPendingMessage" {
		return nil, errors.New("invalid tag: expected WinnersPendingMessage")
	}
	return &WinnersPendingMessage{}, nil
}

func (winnersPendingMessage *WinnersPendingMessage) Serialize() string {
	panic("Not implemented")
}

func (winnersPendingMessage *WinnersPendingMessage) Length() int {
	panic("Not implemented")
}
