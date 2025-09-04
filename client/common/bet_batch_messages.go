package common

import (
	"strings"
)

type BetBatchMessage struct {
	Bets []BetMessage
}

func MatchesBetBatchMessage(s string) bool {
	panic("Not implemented")
}

func DeserializeBetBatchMessage(s string) (*BetBatchMessage, error) {
	panic("Not implemented")
}

func (betBatchMessage *BetBatchMessage) Serialize() string {
	parts := make([]string, 0, len(betBatchMessage.Bets)+1)
	parts = append(parts, "BetBatchMessage")
	for _, bet := range betBatchMessage.Bets {
		parts = append(parts, bet.Serialize())
	}
	return strings.Join(parts, RecordDelimiter)
}

func (betBatchMessage *BetBatchMessage) Length() int {
	return len(betBatchMessage.Serialize())
}
