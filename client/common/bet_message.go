package common

import (
	"fmt"
)

type BetMessage struct {
	Agency    string
	FirstName string
	LastName  string
	Document  string
	Birthdate string
	Number    string
}

func MatchesBetMessage(s string) bool {
	panic("Not implemented")
}

func DeserializeBetMessage(s string) (*BetMessage, error) {
	panic("Not implemented")
}

func (b *BetMessage) Serialize() string {
	return fmt.Sprintf(
		"BetMessage%s%s%s%s%s%s%s%s%s%s%s%s",
		FieldDelimiter, b.Agency,
		FieldDelimiter, b.FirstName,
		FieldDelimiter, b.LastName,
		FieldDelimiter, b.Document,
		FieldDelimiter, b.Birthdate,
		FieldDelimiter, b.Number,
	)
}

func (b *BetMessage) Length() int {
	return len(b.Serialize())
}
