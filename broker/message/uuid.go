package message

import (
	"encoding/json"

	"github.com/twinj/uuid"
)

type UUID struct {
	sub uuid.UUID
}

func NewUUID() *UUID {
	return &UUID{sub: uuid.NewV4()}
}

func ParseUUID(str string) (*UUID, error) {
	ret, err := uuid.Parse(str)
	if err != nil {
		return nil, err
	}
	return &UUID{sub: *ret}, nil
}

func MustUUID(str string) *UUID {
	ret, err := ParseUUID(str)
	if err != nil {
		panic(err)
	}
	return ret
}

func (u UUID) String() string {
	return u.sub.String()
}

func (u UUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.sub.String())
}

func (u *UUID) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(str)
	if err != nil {
		return err
	}
	u.sub = *id
	return nil
}
