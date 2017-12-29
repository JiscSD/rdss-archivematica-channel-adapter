package message

type Organisation struct {
	OrganisationJiscId  int                  `json:"organisationJiscId"`
	OrganisationName    string               `json:"organisationName"`
	OrganisationType    OrganisationTypeEnum `json:"organisationType"`
	OrganisationAddress string               `json:"organisationAddress"`
}

type Person struct {
	PersonUuid            *UUID                            `json:"personUuid"`
	PersonIdentifier      []PersonIdentifier               `json:"personIdentifier"`
	PersonEntitlement     []PersonRoleEnum                 `json:"personEntitlement"`
	PersonOrganisation    []Organisation                   `json:"personOrganisation"`
	PersonAffiliation     []EduPersonScopedAffiliationEnum `json:"personAffiliation"`
	PersonGivenName       string                           `json:"personGivenName"`
	PersonCn              string                           `json:"personCn"`
	PersonSn              string                           `json:"personSn"`
	PersonTelephoneNumber string                           `json:"personTelephoneNumber"`
	PersonMail            string                           `json:"personMail"`
	PersonOu              string                           `json:"personOu"`
}

type PersonIdentifier struct {
	PersonIdentifierValue string                   `json:"personIdentifierValue"`
	PersonIdentifierType  PersonIdentifierTypeEnum `json:"personIdentifierType"`
}
