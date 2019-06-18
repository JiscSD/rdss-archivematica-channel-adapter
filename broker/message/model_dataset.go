package message

// Dataset inherits from ResearchObject.
type Dataset struct {
	ObjectUUID              *UUID                    `json:"objectUUID"`
	ObjectTitle             string                   `json:"objectTitle"`
	ObjectPersonRole        []PersonRole             `json:"objectPersonRole"`
	ObjectDescription       []ObjectDescription      `json:"objectDescription,omitempty"`
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
	Language                []string                 `json:"language,omitempty"`
	Coverage                Coverage                 `json:"coverage,omitempty"`
	Version                 string                   `json:"version,omitempty"`
}
