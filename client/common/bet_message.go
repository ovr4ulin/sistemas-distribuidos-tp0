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

func (b *BetMessage) Serialize() string {
	return fmt.Sprintf(
		"BetMessage^%s^%s^%s^%s^%s^%s",
		b.Agency, b.FirstName, b.LastName, b.Document, b.Birthdate, b.Number,
	)
}

func (b *BetMessage) Length() int {
	return len(b.Serialize())
}
