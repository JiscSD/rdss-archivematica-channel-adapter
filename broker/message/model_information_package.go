package message

type InformationPackage struct {
	ObjectUUID               *UUID              `json:"objectUUID"`
	PackageUUID              *UUID              `json:"packageUUID"`
	PackageType              *PackageTypeEnum   `json:"packageType"`
	PackageContainerType     *ContainerTypeEnum `json:"packageContainerType"`
	PackageDescription       string             `json:"packageDescription,omitempty"`
	PackagePreservationEvent PreservationEvent  `json:"packagePreservationEvent"`
}

type PreservationEvent struct {
	PreservationEventValue  string                     `json:"preservationEventValue"`
	PreservationEventType   *PreservationEventTypeEnum `json:"preservationEventType"`
	PreservationEventDetail string                     `json:"preservationEventDetail,omitempty"`
}
