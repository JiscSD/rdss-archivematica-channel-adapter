package message_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message/specdata"
)

// This is the one message type that we expect to be publishing from the adapter
// in order to describe an event during the preservation of one object in RDSS.
func TestMessagePreservation_Marshal(t *testing.T) {
	m := message.New(message.MessageTypeEnum_PreservationEvent, message.MessageClassEnum_Event)
	b, _ := m.PreservationEventRequest()
	b.InformationPackage = message.InformationPackage{}
	have, err := json.MarshalIndent(m.MessageBody, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	want := []byte(`{
  "objectUUID": null,
  "packageUUID": null,
  "packageType": null,
  "packageContainerType": null,
  "packagePreservationEvent": {
    "preservationEventValue": "",
    "preservationEventType": null
  }
}`)
	if !bytes.Equal(have, want) {
		t.Errorf("have %s, expected %s", have, want)
	}
}

func TestMessagePreservation_Unmarshal(t *testing.T) {
	blob := specdata.MustAsset("messages/body/preservation/preservation_event_request.json")
	packageType := message.PackageTypeEnum_AIP
	packageContainerType := message.ContainerTypeEnum_zip
	preservationEventType := message.PreservationEventTypeEnum_informationPackageCreation
	have := message.PreservationEventRequest{}
	want := message.PreservationEventRequest{
		InformationPackage: message.InformationPackage{
			PackageUUID:          message.MustUUID("5680e8e0-28a5-4b20-948e-fd0d08781e0b"),
			ObjectUUID:           message.MustUUID("0a021cf3-fa5f-4b86-ab09-634bd5de1abd"),
			PackageType:          &packageType,
			PackageContainerType: &packageContainerType,
			PackageDescription:   "A free text string",
			PackagePreservationEvent: message.PreservationEvent{
				PreservationEventValue:  "A free text string",
				PreservationEventType:   &preservationEventType,
				PreservationEventDetail: "A free text string",
			},
		},
	}
	if err := json.Unmarshal(blob, &have); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(have, want) {
		t.Errorf("unmarshal returned an unexpected value; want %v, have %v", want, have)
	}
}
