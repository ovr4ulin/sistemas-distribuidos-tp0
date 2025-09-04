package common

import (
	"fmt"
)

type WinnersRequestMessage struct {
	Agency string
}

func MatchesWinnersRequestMessage(s string) bool {
	panic("Not implemented")
}

func DeserializeWinnersRequestMessage(s string) (*BetMessage, error) {
	panic("Not implemented")
}

func (winnersRequestMessage *WinnersRequestMessage) Serialize() string {
	return fmt.Sprintf("WinnersRequestMessage%s%s", FieldDelimiter, winnersRequestMessage.Agency)
}

func (winnersRequestMessage *WinnersRequestMessage) Length() int {
	return len(winnersRequestMessage.Serialize())
}
