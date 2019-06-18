package message

type File struct {
	FileUUID               *UUID               `json:"fileUUID"`
	FileIdentifier         string              `json:"fileIdentifier"`
	FileName               string              `json:"fileName"`
	FileSize               int                 `json:"fileSize"`
	FileLabel              string              `json:"fileLabel,omitempty"`
	FileDateCreated        *Timestamp          `json:"fileDateCreated,omitempty"`
	FileRights             *Rights             `json:"fileRights,omitempty"`
	FileChecksum           []Checksum          `json:"fileChecksum"`
	FileFormatType         string              `json:"fileFormatType,omitempty"`
	FileCompositionLevel   string              `json:"fileCompositionLevel"`
	FileHasMimeType        bool                `json:"fileHasMimeType,omitempty"`
	FileDateModified       []Timestamp         `json:"fileDateModified"`
	FilePuid               []string            `json:"filePuid,omitempty"`
	FileUse                FileUseEnum         `json:"fileUse"`
	FilePreservationEvent  string              `json:"filePreservationEvent,omitempty"`
	FileUploadStatus       UploadStatusEnum    `json:"fileUploadStatus"`
	FileStorageStatus      StorageStatusEnum   `json:"fileStorageStatus"`
	FileLastDownload       *Timestamp          `json:"fileLastDownloaded,omitempty"`
	FileTechnicalAttribute []string            `json:"fileTechnicalAttribute,omitempty"`
	FileStorageLocation    string              `json:"fileStorageLocation"`
	FileStoragePlatform    FileStoragePlatform `json:"fileStoragePlatform"`
}

type Checksum struct {
	ChecksumUUID  *UUID            `json:"checksumUUID,omitempty"`
	ChecksumType  ChecksumTypeEnum `json:"checksumType"`
	ChecksumValue string           `json:"checksumValue"`
}

type FileStoragePlatform struct {
	StoragePlatformUUID *UUID           `json:"storagePlatformUUID"`
	StoragePlatformName string          `json:"storagePlatformName"`
	StoragePlatformType StorageTypeEnum `json:"storagePlatformType"`
	StoragePlatformCost string          `json:"storagePlatformCost"`
}

type Grant struct {
	GrantUUID       *UUID            `json:"grantUUID"`
	GrantIdentifier string           `json:"grantIdentifier"`
	GrantFunder     OrganisationRole `json:"grantFunder"`
	GrantStart      Timestamp        `json:"grantStart"`
	GrantEnd        Timestamp        `json:"grantEnd"`
}

type Project struct {
	ProjectUUID        *UUID        `json:"projectUUID"`
	ProjectIdentifier  []Identifier `json:"projectIdentifier"`
	ProjectName        string       `json:"projectName"`
	ProjectDescription string       `json:"projectDescription"`
	ProjectCollection  []Collection `json:"projectCollection"`
	ProjectGrant       []Grant      `json:"projectGrant,omitempty"`
	ProjectStart       Timestamp    `json:"projectStart"`
	ProjectEnd         Timestamp    `json:"projectEnd"`
}
