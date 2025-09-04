package common

import (
	"fmt"
)

type EndOfBetsMessage struct {
	Agency string
}

func MatchesEndOfBetsMessage(s string) bool {
	panic("Not implemented")
}

func DeserializeEndOfBetsMessage(s string) (*BetMessage, error) {
	panic("Not implemented")
}

func (endOfBetsMessage *EndOfBetsMessage) Serialize() string {
	return fmt.Sprintf("EndOfBetsMessage%s%s", FieldDelimiter, endOfBetsMessage.Agency)
}

func (endOfBetsMessage *EndOfBetsMessage) Length() int {
	return len(endOfBetsMessage.Serialize())
}
