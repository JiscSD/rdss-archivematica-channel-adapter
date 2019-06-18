package message

// ThesisDissertation inherits from ResearchObject.
type ThesisDissertation struct {
	ObjectUUID              *UUID                                 `json:"objectUUID"`
	ObjectTitle             string                                `json:"objectTitle"`
	ObjectPersonRole        []PersonRole                          `json:"objectPersonRole"`
	ObjectDescription       []ObjectDescription                   `json:"objectDescription"`
	ObjectRights            Rights                                `json:"objectRights"`
	ObjectDate              []ThesisAwardedDate                   `json:"objectDate"`
	ObjectKeyword           []string                              `json:"objectKeyword,omitempty"`
	ObjectCategory          []string                              `json:"objectCategory,omitempty"`
	ObjectResourceType      ResourceTypeEnum                      `json:"objectResourceType"`
	ObjectValue             ObjectValueEnum                       `json:"objectValue"`
	ObjectIdentifier        []Identifier                          `json:"objectIdentifier"`
	ObjectRelatedIdentifier []IdentifierRelationship              `json:"objectRelatedIdentifier,omitempty"`
	ObjectOrganisationRole  []AwardingInstitutionOrganisationRole `json:"objectOrganisationRole"`
	ObjectFile              []File                                `json:"objectFile,omitempty"`
	Coverage                Coverage                              `json:"coverage,omitempty"`
	Language                []string                              `json:"language"`
	QualificationLevel      string                                `json:"qualificationLevel"`
	QualificationName       string                                `json:"qualificationName"`
}

type ThesisAwardedDate struct {
	DateValue Timestamp    `json:"dateValue"`
	DateType  DateTypeEnum `json:"dateType"`
}

type AwardingInstitutionOrganisationRole struct {
	Organisation Organisation         `json:"organisation"`
	Role         OrganisationRoleEnum `json:"role"`
}
