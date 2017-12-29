package message

import (
	"bytes"
	"encoding/json"
	"testing"
)

const (
	validUUID   = "be8eff14-a92b-429e-80b8-0ec4594d72c0"
	invalidUUID = "--invalid-uuid--"
)

func TestParseUUID(t *testing.T) {
	if _, err := ParseUUID(validUUID); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseUUID(invalidUUID); err == nil {
		t.Fatal(err)
	}
}

func TestMustUUIDWithPanic(t *testing.T) {
	t.Run("WithPanic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("The code did not panic")
			}
		}()
		MustUUID(invalidUUID)
	})
	t.Run("WithoutPanic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Error("The code did panic")
			}
		}()
		MustUUID(validUUID)
	})
}

func TestMarshalJSON(t *testing.T) {
	in := NewUUID()
	have, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	const quote = byte(rune('"'))
	want := []byte(in.String())
	want = append([]byte{quote}, want...)
	want = append(want, quote)

	if !bytes.Equal(have, want) {
		t.Fatalf("Unexpected result, have %v, want %v", have, want)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name       string
		uuid       string
		shouldFail bool
	}{
		{"Valid UUID", validUUID, false},
		{"Invalid UUID", invalidUUID, true},
		{"Invalid JSON", `", "`, true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				encoded = []byte(`"` + tc.uuid + `"`)
				id      = &UUID{}
			)
			err := json.Unmarshal(encoded, id)
			if tc.shouldFail && err == nil {
				t.Fatal("Decoding did not fail.")
			} else if !tc.shouldFail && err != nil {
				t.Fatalf("Decoding failed: %v", err)
			}
			if tc.shouldFail {
				return
			}
			if have, want := id.String(), tc.uuid; have != want {
				t.Fatalf("Unexpected output; have %s, want %s.", have, want)
			}
		})
	}
}
