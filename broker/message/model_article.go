package message

// Article inherits from ResearchObject.
type Article struct {
	ObjectUUID              *UUID                    `json:"objectUUID"`
	ObjectTitle             string                   `json:"objectTitle"`
	ObjectPersonRole        []PersonRole             `json:"objectPersonRole"`
	ObjectDescription       []ObjectDescription      `json:"objectDescription"`
	ObjectRights            Rights                   `json:"objectRights"`
	ObjectDate              []Date                   `json:"objectDate"`
	ObjectKeyword           []string                 `json:"objectKeyword,omitempty"`
	ObjectCategory          []string                 `json:"objectCategory,omitempty"`
	ObjectResourceType      ResourceTypeEnum         `json:"objectResourceType"`
	ObjectValue             ObjectValueEnum          `json:"objectValue"`
	ObjectIdentifier        []Identifier             `json:"objectIdentifier"`
	ObjectRelatedIdentifier []IdentifierRelationship `json:"objectRelatedIdentifier,omitempty"`
	ObjectOrganisationRole  []OrganisationRole       `json:"objectOrganisationRole"`
	ObjectFile              []File                   `json:"objectFile,omitempty"`
	FreeToRead              ALIFreeToRead            `json:"freeToRead,omitempty"`
	Coverage                Coverage                 `json:"coverage,omitempty"`
	Language                []string                 `json:"language"`
	ArticleProcessingCharge RIOXXTermsAPCEnum        `json:"articleProcessingCharge,omitempty"`
	PublicationVersion      PublicationVersionEnum   `json:"publicationVersion"`
	Journal                 Journal                  `json:"journal,omitempty"`
}

type ALIFreeToRead struct {
	StartDate Timestamp `json:"startDate,omitempty"`
	EndDate   Timestamp `json:"endDate,omitempty"`
}

type Journal struct {
	ISSN          string `json:"ISSN"`
	FullTitle     string `json:"fullTitle,omitempty"`
	JournalVolume string `json:"journalVolume"`
	FirstPage     string `json:"firstPage"`
	LastPage      string `json:"lastPage"`
	JournalIssue  string `json:"journalIssue"`
}
