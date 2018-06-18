package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

type Doc struct {
	Prop Timestamp `json:"prop,omitempty"`
}

func TestTimestampEncoding(t *testing.T) {
	testCases := []struct {
		doc      *Doc
		expected []byte
	}{
		{
			&Doc{Prop: Timestamp(time.Date(1997, time.July, 16, 19, 20, 0, 0, time.FixedZone("+0100", 3600)))},
			[]byte(`{"prop":"1997-07-16T19:20:00+01:00"}`),
		},
		{
			&Doc{Prop: Timestamp(time.Date(1997, time.July, 16, 19, 20, 30, 0, time.FixedZone("+0100", 3600)))},
			[]byte(`{"prop":"1997-07-16T19:20:30+01:00"}`),
		},
		{
			&Doc{Prop: Timestamp(time.Date(1997, time.July, 16, 19, 20, 30, 450000000, time.FixedZone("+0100", 3600)))},
			[]byte(`{"prop":"1997-07-16T19:20:30.45+01:00"}`),
		},
		{
			&Doc{Prop: Timestamp(time.Date(2004, time.August, 1, 10, 0, 0, 0, time.UTC))},
			[]byte(`{"prop":"2004-08-01T10:00:00-00:00"}`),
		},
		{
			&Doc{Prop: Timestamp{}},
			[]byte(`{"prop":null}`),
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Encoding %s", tc.doc.Prop), func(t *testing.T) {
			data, err := json.Marshal(tc.doc)
			if err != nil {
				t.Errorf("error encoding: %s", err)
			}
			if !bytes.Equal(data, tc.expected) {
				t.Errorf("unexpected value:\nhave: %s;\nexpected: %s", data, tc.expected)
			}
		})
	}
}

func TestTimestampDecoding(t *testing.T) {
	testCases := []struct {
		data     []byte
		format   string
		expected time.Time
		wantErr  bool
	}{
		{
			[]byte(`{"prop": "1997-07-16T19:20+01:00"}`),
			"YYYY-MM-DDThh:mmTZD",
			time.Date(1997, time.July, 16, 19, 20, 0, 0, time.FixedZone("+0100", 3600)),
			false,
		},
		{
			[]byte(`{"prop": "1997-07-16T19:20:30+01:00"}`),
			"YYYY-MM-DDThh:mm:ssTZD",
			time.Date(1997, time.July, 16, 19, 20, 30, 0, time.FixedZone("+0100", 3600)),
			false,
		},
		{
			[]byte(`{"prop": "1997-07-16T19:20:30.45+01:00"}`),
			"YYYY-MM-DDThh:mm:ss.sTZD",
			time.Date(1997, time.July, 16, 19, 20, 30, 450000000, time.FixedZone("+0100", 3600)),
			false,
		},
		{
			[]byte(`{"prop": null}`),
			"Zero value",
			time.Time{},
			false,
		},
		{
			[]byte(`{"prop": ""}`),
			"Empty string",
			time.Time{},
			false,
		},
		{
			[]byte(`{"prop": "2006-01-02TZ"}`),
			"Unknown format",
			time.Time{},
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Decoding format %s", tc.format), func(t *testing.T) {
			target := &Doc{}
			err := json.Unmarshal(tc.data, &target)
			if tc.wantErr {
				if err == nil {
					t.Errorf("should return an error")
				}
				return
			}
			if err != nil {
				t.Errorf("error decoding: %s", err)
			}
			if !time.Time(target.Prop).Equal(tc.expected) {
				t.Errorf("unexpected value:\nhave: %s;\nexpected: %s", target.Prop, tc.expected)
			}
		})
	}
}
