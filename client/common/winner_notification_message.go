package common

import (
	"errors"
	"strconv"
	"strings"
)

type WinnersNotificationMessage struct {
	Count     int
	Documents []int
}

func MatchesWinnersNotificationMessage(s string) bool {
	return matchTag(s, "WinnersNotificationMessage")
}

func DeserializeWinnersNotificationMessage(s string) (*WinnersNotificationMessage, error) {
	parts := strings.Split(s, FieldDelimiter)
	if len(parts) != 3 || parts[0] != "WinnersNotificationMessage" {
		return nil, errors.New("invalid WinnersNotificationMessage format")
	}

	count, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, errors.New("invalid count in WinnersNotificationMessage")
	}

	docs := []int{}
	if parts[2] != "" {
		raw := strings.Split(parts[2], RecordDelimiter)
		docs = make([]int, 0, len(raw))
		for _, r := range raw {
			n, err := strconv.Atoi(r)
			if err != nil {
				return nil, errors.New("invalid document in WinnersNotificationMessage: " + r)
			}
			docs = append(docs, n)
		}
	}

	if count != len(docs) {
		return nil, errors.New("count does not match number of documents")
	}

	return &WinnersNotificationMessage{
		Count:     count,
		Documents: docs,
	}, nil
}

func (m *WinnersNotificationMessage) Serialize() string {
	panic("Not implemented")
}

func (m *WinnersNotificationMessage) Length() int {
	panic("Not implemented")
}
