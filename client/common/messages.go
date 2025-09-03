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

func NewBetMessage(agency string, firstName string, lastName string, document string, birthdate string, number string) *BetMessage {
	return &BetMessage{
		Agency:    agency,
		FirstName: firstName,
		LastName:  lastName,
		Document:  document,
		Birthdate: birthdate,
		Number:    number,
	}
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
		b.Agency,
		b.FirstName,
		b.LastName,
		b.Document,
		b.Birthdate,
		b.Number,
	)
}

func (b *BetMessage) MatchesBetMessage(s string) bool {
	return matchTag(s, "BetMessage")
}

func (b *BetMessage) Length() int {
	return len(b.Serialize())
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
	return &AckMessage{
		ProcessedCount: parts[1],
	}, nil
}

func matchTag(s, expected string) bool {
	parts := strings.SplitN(s, "^", 2)
	return len(parts) > 0 && parts[0] == expected
}
