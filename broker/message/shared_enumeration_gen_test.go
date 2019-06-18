package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestEnumerationGenerator(t *testing.T) {
	var uploadStatus UploadStatusEnum = UploadStatusEnum_uploadStarted
	if have, want := "uploadStarted", fmt.Sprint(uploadStatus); have != want {
		t.Fatalf("UploadStatusEnum.String() returned an unexpected value; want %v, have %v", want, have)
	}

	type data struct {
		UploadStatus UploadStatusEnum
	}

	// Test decoding
	blob := []byte(`{"uploadStatus": "uploadStarted"}`)
	var doc = data{}
	if err := json.Unmarshal(blob, &doc); err != nil {
		t.Fatal(err)
	}
	if have, want := UploadStatusEnum_uploadStarted, uploadStatus; have != want {
		t.Fatalf("UploadStatusEnum was not decoded as expected; want %v, have %v", want, have)
	}

	// Test encoding
	doc = data{UploadStatus: UploadStatusEnum_uploadStarted}
	blob, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if have, want := blob, []byte(`{"UploadStatus":"uploadStarted"}`); !bytes.Equal(have, want) {
		t.Fatalf("UploadStatusEnum was not decoded as expected; want %s, have %s", want, have)
	}
}
