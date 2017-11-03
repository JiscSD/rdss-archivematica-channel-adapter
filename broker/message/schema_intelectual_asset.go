package message

type File struct {
	FileUUID                string              `json:"fileUuid"`
	FileIdentifier          string              `json:"fileIdentifier"`
	FileName                string              `json:"fileName"`
	FileSize                int                 `json:"fileSize"`
	FileLabel               string              `json:"fileLabel,omitempty"`
	FileDateCreated         Date                `json:"fileDateCreated,omitempty"`
	FileRights              Rights              `json:"fileRights,omitempty"`
	FileChecksum            []Checksum          `json:"fileChecksum"`
	FileFormatType          string              `json:"fileFormatType,omitempty"`
	FileCompositionLevel    string              `json:"fileCompositionLevel"`
	FileHasMimeType         bool                `json:"fileHasMimeType,omitempty"`
	FileDateModified        []Date              `json:"fileDateModified"`
	FilePuid                []string            `json:"filePuid,omitempty"`
	FileUse                 FileUseEnum         `json:"fileUse"`
	FilePreservationEvent   []PreservationEvent `json:"filePreservationEvent"`
	FileUploadStatus        UploadStatusEnum    `json:"fileUploadStatus"`
	FileStorageStatus       StorageStatusEnum   `json:"fileStorageStatus"`
	FileLastDownload        Date                `json:"fileLastDownloaded,omitempty"`
	FileTechnicalAttributes []string            `json:"fileTechnicalAttributes,omitempty"`
	FileStorageLocation     string              `json:"fileStorageLocation"`
	FileStorageType         StorageTypeEnum     `json:"fileStorageType"`
}

type Checksum struct {
	ChecksumUuid  string           `json:"checksumUuid,omitempty"`
	ChecksumType  ChecksumTypeEnum `json:"checksumType"`
	ChecksumValue string           `json:"checksumValue"`
}

type PreservationEvent struct {
	PreservationEventValue  string                    `json:"preservationEventValue"`
	PreservationEventType   PreservationEventTypeEnum `json:"preservationEventType"`
	PreservationEventDetail string                    `json:"preservationEventDetail,omitempty"`
}

type Permission struct {
	Read    bool `json:"read"`
	Write   bool `json:"write"`
	Control bool `json:"control"`
	Append  bool `json:"append"`
}

type FilePermission struct {
	Permission Permission `json:"permission"`
	File       File       `json:"File"`
}

type Group struct {
	GroupUuid           string           `json:"groupUuid"`
	GroupName           string           `json:"groupName"`
	GroupIdentifier     string           `json:"groupIdentifier"`
	GroupFilePermission []FilePermission `json:"groupFilePermission"`
	GroupMembers        []Person         `json:"groupMembers,omitempty"`
}

type Grant struct {
	GrantUuid       string           `json:"grantUuid"`
	GrantIdentifier string           `json:"grantIdentifier"`
	GrantFunder     OrganisationRole `json:"grantFunder"`
	GrantStart      Date             `json:"grantStart"`
	GrantEnd        Date             `json:"grantEnd"`
	GrantValue      string           `json:"grantValue"`
}

type Project struct {
	ProjectUuid        string       `json:"projectUuid"`
	ProjectIdentifier  []string     `json:"projectIdentifier"`
	ProjectName        string       `json:"projectName"`
	ProjectDescription string       `json:"projectDescription"`
	ProjectCollection  []Collection `json:"projectCollection"`
	ProjectGroup       []Group      `json:"projectGroup"`
	ProjectGrant       []Grant      `json:"projectGrant,omitempty"`
	ProjectStart       Date         `json:"projectStart"`
	ProjectEnd         Date         `json:"projectEnd"`
}
