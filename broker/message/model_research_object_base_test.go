package message_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
)

func TestResearchObjectBase_ToJSON(t *testing.T) {
	testCases := []message.ResearchObjectBase{
		message.ResearchObjectBase{Article: &message.Article{}},
		message.ResearchObjectBase{Dataset: &message.Dataset{}},
		message.ResearchObjectBase{ThesisDissertation: &message.ThesisDissertation{}},
		message.ResearchObjectBase{ResearchObject: &message.ResearchObject{}},
	}
	for _, tt := range testCases {
		_, err := json.Marshal(tt)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestResearchObjectBase_FromJSON(t *testing.T) {
	testCases := []struct {
		payload []byte
		prop    string
	}{
		{
			[]byte(`{
				"objectUUID": "e01d51f5-57e7-422a-9622-448eaef6abf3",
				"objectResourceType": "article"
			}`),
			"Article",
		},
		{
			[]byte(`{
				"objectUUID": "e01d51f5-57e7-422a-9622-448eaef6abf3",
				"objectResourceType": "dataset"
			}`),
			"Dataset",
		},
		{
			[]byte(`{
				"objectUUID": "e01d51f5-57e7-422a-9622-448eaef6abf3",
				"objectResourceType": "thesisDissertation"
			}`),
			"ThesisDissertation",
		},
		{
			[]byte(`{
				"objectUUID": "e01d51f5-57e7-422a-9622-448eaef6abf3",
				"objectResourceType": "other"
			}`),
			"ResearchObject",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.prop, func(t *testing.T) {
			v := message.ResearchObjectBase{}
			if err := json.Unmarshal(tt.payload, &v); err != nil {
				t.Fatal(err)
			}
			r := reflect.ValueOf(&v).Elem()
			for i := 0; i < r.NumField(); i++ {
				field := r.Field(i)
				name := field.Type().String()
				if name == fmt.Sprintf("*message.%s", tt.prop) {
					if field.IsNil() {
						t.Fatalf("%s should not be nil", name)
					}
					if tname := reflect.Indirect(field).Type().Name(); tname != tt.prop {
						t.Fatalf("%s has a non-nil value but it is a %s but a %s", name, tname, tt.prop)
					}
					uuid := fmt.Sprintf("%s", reflect.Indirect(field).FieldByName("ObjectUUID"))
					if uuid != "e01d51f5-57e7-422a-9622-448eaef6abf3" {
						t.Fatalf("objectUUID seen after unserialization unexpected %s", uuid)
					}
				}
				if name != fmt.Sprintf("*message.%s", tt.prop) && !field.IsNil() {
					t.Fatalf("%s should be nil", name)
				}
			}
		})
	}
}

func TestResearchObjectBase_InferResearchObject(t *testing.T) {
	base := message.ResearchObjectBase{Article: &message.Article{
		ObjectTitle: "title",
	}}
	ro := base.InferResearchObject()
	if ro.ObjectTitle != "title" {
		t.Fatal("title is not defined properly")
	}
}
