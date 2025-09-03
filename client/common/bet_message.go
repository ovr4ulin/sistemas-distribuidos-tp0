package common

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type BetMessage struct {
	Agency    string
	FirstName string
	LastName  string
	Document  string
	Birthdate string
	Number    string
}

func NewBetMessage(agency, firstName, lastName, document, birthdate, number string) *BetMessage {
	return &BetMessage{agency, firstName, lastName, document, birthdate, number}
}

func NewBetMessageFromEnv() *BetMessage {
	return &BetMessage{
		Agency:    os.Getenv("CLI_ID"),
		FirstName: os.Getenv("FIRST_NAME"),
		LastName:  os.Getenv("LAST_NAME"),
		Document:  os.Getenv("DOCUMENT"),
		Birthdate: os.Getenv("BIRTH_DATE"),
		Number:    os.Getenv("NUMBER"),
	}
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

func MatchesBetMessage(s string) bool {
	return matchTag(s, "BetMessage")
}

func DeserializeBetMessage(s string) (*BetMessage, error) {
	parts := strings.Split(s, "^")
	if len(parts) != 7 {
		return nil, errors.New("invalid BetMessage format")
	}
	if parts[0] != "BetMessage" {
		return nil, errors.New("invalid tag: expected BetMessage")
	}
	return &BetMessage{
		Agency:    parts[1],
		FirstName: parts[2],
		LastName:  parts[3],
		Document:  parts[4],
		Birthdate: parts[5],
		Number:    parts[6],
	}, nil
}
