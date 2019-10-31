package message

import (
	"strings"
	"time"
)

// Timestamp is similar to time.Time but implementing the formatting specifics
// described in the API: see https://git.io/v5obt for more details.
type Timestamp time.Time

func (t Timestamp) MarshalJSON() ([]byte, error) {
	ts := time.Time(t)
	if ts.IsZero() {
		return []byte("null"), nil
	}
	bytes, err := time.Time(t).MarshalJSON()
	if err != nil {
		return nil, err
	}

	// This is here so we return `2004-08-01T10:00:00-00:00` instead of
	// `2004-08-01T10:00:00Z` just so we match the way they are represented in
	// the examples of the API repo.
	str := string(bytes)
	const suffix = "Z\""
	if strings.HasSuffix(str, suffix) {
		str = str[0:len(str)-len(suffix)] + "-00:00\""
	}

	return []byte(str), nil
}

var timeOtherFormats = []string{
	// Format described in the spec that time.RFC3339 does not cover.
	"2006-01-02T15:04Z07:00",
	// Format seen in the wild: "2019-10-31T16:20:05.921+0000".
	"2006-01-02T15:04:05.999-0700",
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The time is expected to be a quoted string in RFC 3339 format.
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	str := string(data)
	// Ignore null, like in the main JSON package.
	if str == "null" {
		return nil
	}
	// Ignore empty string.
	if str == "\"\"" {
		return nil
	}
	// Fractional seconds are handled implicitly by Parse.
	ts, err := time.Parse(`"`+time.RFC3339+`"`, str)
	if err == nil {
		*t = Timestamp(ts)
		return nil
	}
	// Other formats.
	for _, format := range timeOtherFormats {
		ts, err = time.Parse(`"`+format+`"`, str)
		if err == nil {
			*t = Timestamp(ts)
			return nil
		}
	}
	return err
}

func (t Timestamp) String() string {
	return time.Time(t).String()
}
