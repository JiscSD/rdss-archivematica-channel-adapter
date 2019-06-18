package message

type Organisation struct {
	OrganisationJiscId  int                  `json:"organisationJiscId"`
	OrganisationName    string               `json:"organisationName"`
	OrganisationType    OrganisationTypeEnum `json:"organisationType,omitempty"`
	OrganisationAddress string               `json:"organisationAddress,omitempty"`
}

type Person struct {
	PersonUUID             *UUID              `json:"personUUID"`
	PersonIdentifier       []PersonIdentifier `json:"personIdentifier"`
	PersonHonorificPrefix  string             `json:"personHonorificPrefix,omitempty"`
	PersonGivenNames       string             `json:"personGivenNames"`
	PersonFamilyNames      string             `json:"personFamilyNames"`
	PersonHonorificSuffix  string             `json:"personHonorificSuffix,omitempty"`
	PersonMail             string             `json:"personMail,omitempty"`
	PersonOrganisationUnit *OrganisationUnit  `json:"personOrganisationUnit,omitempty"`
}

type PersonIdentifier struct {
	PersonIdentifierValue string                   `json:"personIdentifierValue"`
	PersonIdentifierType  PersonIdentifierTypeEnum `json:"personIdentifierType"`
}

type OrganisationUnit struct {
	OrganisationUnitUUID *UUID        `json:"organisationUnitUUID"`
	OrganisationUnitName string       `json:"organisationUnitName"`
	Organisation         Organisation `json:"organisation"`
}
