package message

type File struct {
	FileUUID                *UUID               `json:"fileUuid"`
	FileIdentifier          string              `json:"fileIdentifier"`
	FileName                string              `json:"fileName"`
	FileSize                int                 `json:"fileSize"`
	FileLabel               string              `json:"fileLabel,omitempty"`
	FileDateCreated         Timestamp           `json:"fileDateCreated,omitempty"`
	FileRights              Rights              `json:"fileRights,omitempty"`
	FileChecksum            []Checksum          `json:"fileChecksum"`
	FileFormatType          string              `json:"fileFormatType,omitempty"`
	FileCompositionLevel    string              `json:"fileCompositionLevel"`
	FileHasMimeType         bool                `json:"fileHasMimeType,omitempty"`
	FileDateModified        []Timestamp         `json:"fileDateModified"`
	FilePuid                []string            `json:"filePuid,omitempty"`
	FileUse                 FileUseEnum         `json:"fileUse"`
	FilePreservationEvent   []PreservationEvent `json:"filePreservationEvent"`
	FileUploadStatus        UploadStatusEnum    `json:"fileUploadStatus"`
	FileStorageStatus       StorageStatusEnum   `json:"fileStorageStatus"`
	FileLastDownload        Timestamp           `json:"fileLastDownloaded,omitempty"`
	FileTechnicalAttributes []string            `json:"fileTechnicalAttributes,omitempty"`
	FileStorageLocation     string              `json:"fileStorageLocation"`
	FileStoragePlatform     FileStoragePlatform `json:"fileStoragePlatform"`
}

type Checksum struct {
	ChecksumUuid  *UUID            `json:"checksumUuid,omitempty"`
	ChecksumType  ChecksumTypeEnum `json:"checksumType"`
	ChecksumValue string           `json:"checksumValue"`
}

type PreservationEvent struct {
	PreservationEventValue  string                    `json:"preservationEventValue"`
	PreservationEventType   PreservationEventTypeEnum `json:"preservationEventType"`
	PreservationEventDetail string                    `json:"preservationEventDetail,omitempty"`
}

type FileStoragePlatform struct {
	StoragePlatformUuid *UUID           `json:"storagePlatformUuid"`
	StoragePlatformName string          `json:"storagePlatformName"`
	StoragePlatformType StorageTypeEnum `json:"storagePlatformType"`
	StoragePlatformCost string          `json:"storagePlatformCost"`
}

type Grant struct {
	GrantUuid       *UUID            `json:"grantUuid"`
	GrantIdentifier string           `json:"grantIdentifier"`
	GrantFunder     OrganisationRole `json:"grantFunder"`
	GrantStart      Timestamp        `json:"grantStart"`
	GrantEnd        Timestamp        `json:"grantEnd"`
}

type Project struct {
	ProjectUuid        *UUID        `json:"projectUuid"`
	ProjectIdentifier  []Identifier `json:"projectIdentifier"`
	ProjectName        string       `json:"projectName"`
	ProjectDescription string       `json:"projectDescription"`
	ProjectCollection  []Collection `json:"projectCollection"`
	ProjectGrant       []Grant      `json:"projectGrant,omitempty"`
	ProjectStart       Timestamp    `json:"projectStart"`
	ProjectEnd         Timestamp    `json:"projectEnd"`
}
