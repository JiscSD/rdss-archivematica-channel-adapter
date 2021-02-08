package adapter

import (
	"encoding/csv"
	"testing"

	"github.com/spf13/afero"

	"github.com/JiscSD/rdss-archivematica-channel-adapter/amclient"
	"github.com/JiscSD/rdss-archivematica-channel-adapter/broker/message"
)

var md = describeDataset

type mdTest struct {
	objectTitle         string
	creatorType         message.PersonRoleEnum
	creatorGivenName    string
	creatorFamilyName   string
	publisherType       message.PersonRoleEnum
	publisherGivenName  string
	publisherFamilyName string
	identifierType      message.IdentifierTypeEnum
	identifierValue     string
	expected            []string
}

var mdSlice = []mdTest{
	// No family name.
	mdTest{
		"ObjectTitle1",
		message.PersonRoleEnum_dataCreator,
		"Kat",
		"",
		message.PersonRoleEnum_publisher,
		"Joan",
		"",
		message.IdentifierTypeEnum_DOI,
		"10.5072/FK2/QAWS8O",
		[]string{"objects/", "Kat", "10.5072/FK2/QAWS8O", "Joan", "ObjectTitle1", "artDesignItem"},
	},
	// No given name.
	mdTest{
		"ObjectTitle2",
		message.PersonRoleEnum_dataCreator,
		"",
		"Winter",
		message.PersonRoleEnum_publisher,
		"",
		"Watson",
		message.IdentifierTypeEnum_DOI,
		"10.5072/FK2/QAWS81",
		[]string{"objects/", "Winter", "10.5072/FK2/QAWS81", "Watson", "ObjectTitle2", "artDesignItem"},
	},
	// Both names.
	mdTest{
		"ObjectTitle3",
		message.PersonRoleEnum_dataCreator,
		"Kat",
		"Winter",
		message.PersonRoleEnum_publisher,
		"Joan",
		"Watson",
		message.IdentifierTypeEnum_DOI,
		"10.5072/FK2/QAWS82",
		[]string{"objects/", "Winter, Kat", "10.5072/FK2/QAWS82", "Watson, Joan", "ObjectTitle3", "artDesignItem"},
	},
	// No matching enums.
	mdTest{
		"ObjectTitle4",
		message.PersonRoleEnum_editor,
		"Kat",
		"Winter",
		message.PersonRoleEnum_other,
		"Joan",
		"Watson",
		message.IdentifierTypeEnum_DOI,
		"10.5072/FK2/QAWS83",
		[]string{"objects/", "10.5072/FK2/QAWS83", "ObjectTitle4", "artDesignItem"},
	},
}

// makeMDSet will create a metadata set from a test data object.
func makeMDSet(value *mdTest, fs *afero.Afero) *amclient.MetadataSet {
	transferSession := amclient.TransferSession{}
	researchObject := message.ResearchObject{}

	metadataSet := amclient.NewMetadataSet(fs)

	transferSession.Metadata = metadataSet

	researchObject.ObjectTitle = value.objectTitle

	person := message.Person{}
	person.PersonGivenNames = value.creatorGivenName
	person.PersonFamilyNames = value.creatorFamilyName

	personRole := message.PersonRole{}
	personRole.Role = value.creatorType
	personRole.Person = person

	researchObject.ObjectPersonRole = append(researchObject.ObjectPersonRole, personRole)

	person = message.Person{}
	person.PersonGivenNames = value.publisherGivenName
	person.PersonFamilyNames = value.publisherFamilyName

	personRole = message.PersonRole{}
	personRole.Role = value.publisherType
	personRole.Person = person

	researchObject.ObjectPersonRole = append(researchObject.ObjectPersonRole, personRole)

	identifers := []message.Identifier{}

	identifier := message.Identifier{}
	identifier.IdentifierValue = value.identifierValue
	identifier.IdentifierType = message.IdentifierTypeEnum_DOI

	identifers = append(identifers, identifier)

	researchObject.ObjectIdentifier = identifers

	md(&transferSession, &researchObject)

	return metadataSet
}

// TestMetadataGeneration inspects some of the internal functionality
// of the channel adapter and ensures that from a given metadata set
// the correct fields are mapped into an object which then returns an
// expected CSV configuration for Archivematica. The test round-trips
// the CSV generation and checks that when it is consumed the row data
// is as is expected.
func TestMetadataGeneration(t *testing.T) {

	const mdFile string = "/metadata/metadata.csv"

	for _, value := range mdSlice {

		fs := afero.Afero{Fs: afero.NewBasePathFs(afero.NewMemMapFs(), "/")}

		metadataSet := makeMDSet(&value, &fs)

		metadataSet.Write()

		csvFile, err := fs.Open(mdFile)
		if err != nil {
			t.Fatal(err)
		}

		csvObject := csv.NewReader(csvFile)

		// Read the first line so we can move onto the row data.
		_, err = csvObject.Read()
		if err != nil {
			t.Errorf("Cannot read CSV file")
		}

		// Read the second line, the row data.
		record, err := csvObject.Read()
		if err != nil {
			t.Errorf("Cannot read CSV file")
		}

		if len(record) != len(value.expected) {
			t.Errorf(
				"CSV record not correct length: '%d', expected: '%d', %s",
				len(record),
				len(value.expected),
				value.expected,
			)
			continue
		}
		for idx := range record {
			if record[idx] != value.expected[idx] {
				t.Errorf(
					"Incorrect value in CSV record: %s %s",
					record[idx],
					value.expected[idx],
				)
				break
			}
		}
	}
}
