package message

type ResearchObject struct {
	ObjectUuid              *UUID               `json:"objectUuid"`
	ObjectTitle             string              `json:"objectTitle"`
	ObjectPersonRole        []PersonRole        `json:"objectPersonRole"`
	ObjectDescription       string              `json:"objectDescription"`
	ObjectRights            []Rights            `json:"objectRights"`
	ObjectDate              []Date              `json:"objectDate"`
	ObjectKeywords          []string            `json:"objectKeywords,omitempty"`
	ObjectCategory          []string            `json:"objectCategory,omitempty"`
	ObjectResourceType      ResourceTypeEnum    `json:"objectResourceType"`
	ObjectValue             ObjectValueEnum     `json:"objectValue"`
	ObjectIdentifier        []Identifier        `json:"objectIdentifier"`
	ObjectRelatedIdentifier []Identifier        `json:"objectRelatedIdentifier,omitempty"`
	ObjectOrganisationRole  []OrganisationRole  `json:"objectOrganisationRole,omitempty"`
	ObjectPreservationEvent []PreservationEvent `json:"objectPreservationEvent,omitempty"`
	ObjectFile              []File              `json:"objectFile,omitempty"`
}

type PersonRole struct {
	Person *Person        `json:"person"`
	Role   PersonRoleEnum `json:"role"`
}

type Rights struct {
	RightsStatement []string  `json:"rightsStatement"`
	RightsHolder    []string  `json:"rightsHolder"`
	Licence         []Licence `json:"licence"`
	Access          []Access  `json:"access"`
}

type Licence struct {
	LicenceName       string `json:"licenceName"`
	LicenceIdentifier string `json:"licenceIdentifier"`
}

type Access struct {
	AccessType      AccessTypeEnum `json:"accessType"`
	AccessStatement string         `json:"accessStatement,omitempty"`
}

type Date struct {
	DateValue string       `json:"dateValue"`
	DateType  DateTypeEnum `json:"dateType"`
}

type Identifier struct {
	IdentifierValue string             `json:"identifierValue"`
	IdentifierType  IdentifierTypeEnum `json:"identifierType"`
	RelationType    RelationTypeEnum   `json:"relationType,omitempty"`
}

type Collection struct {
	CollectionUuid              *UUID              `json:"collectionUuid"`
	CollectionName              string             `json:"collectionName"`
	CollectionObject            []ResearchObject   `json:"collectionObject,omitempty"`
	CollectionKeywords          []string           `json:"collectionKeywords,omitempty"`
	CollectionCategory          []string           `json:"collectionCategory,omitempty"`
	CollectionDescription       []string           `json:"collectionDescription"`
	CollectionRights            []Rights           `json:"collectionRights"`
	CollectionIdentifier        []Identifier       `json:"collectionIdentifier,omitempty"`
	CollectionRelatedIdentifier []Identifier       `json:"collectionRelatedIdentifier,omitempty"`
	CollectionPersonRole        []PersonRole       `json:"collectionPersonRole,omitempty"`
	CollectionOrganisationRole  []OrganisationRole `json:"collectionOrganisationRole,omitempty"`
}

type OrganisationRole struct {
	Organisation *Organisation        `json:"organisation"`
	Role         OrganisationRoleEnum `json:"role"`
}
