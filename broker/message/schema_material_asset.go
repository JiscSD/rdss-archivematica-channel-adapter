package message

type Organisation struct {
	OrganisationJiscId  int                  `json:"organisationJiscId"`
	OrganisationName    string               `json:"organisationName"`
	OrganisationType    OrganisationTypeEnum `json:"organisationType"`
	OrganisationAddress string               `json:"organisationAddress"`
}

type Person struct {
	PersonUuid             *UUID              `json:"personUuid"`
	PersonIdentifier       []PersonIdentifier `json:"personIdentifier"`
	PersonHonorificPrefix  string             `json:"personHonorificPrefix,omitempty"`
	PersonGivenNames       string             `json:"personGivenNames"`
	PersonFamilyNames      string             `json:"personFamilyNames"`
	PersonHonorificSuffix  string             `json:"personHonorificSuffix,omitempty"`
	PersonMail             string             `json:"personMail,omitempty"`
	PersonOrganisationUnit OrganisationUnit   `json:"personOrganisationUnit"`
}

type PersonIdentifier struct {
	PersonIdentifierValue string                   `json:"personIdentifierValue"`
	PersonIdentifierType  PersonIdentifierTypeEnum `json:"personIdentifierType"`
}

type OrganisationUnit struct {
	OrganisationUnitUuid *UUID        `json:"organisationUnitUuid"`
	OrganisationUuidName string       `json:"organisationUnitName"`
	Organisation         Organisation `json:"organisation"`
}
