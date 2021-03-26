package adapter

import (
	"encoding/csv"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/JiscSD/rdss-archivematica-channel-adapter/amclient"
	"github.com/JiscSD/rdss-archivematica-channel-adapter/broker/message"
)

var metadataGenerationTests = map[string]struct {
	researchObject message.ResearchObject // Given this research object,
	expected       []string               // CSV record that we expect, as a slice of strings.
}{
	"No family name": {
		researchObject: message.ResearchObject{
			ObjectTitle: "ObjectTitle1",
			ObjectPersonRole: []message.PersonRole{
				{
					Role: message.PersonRoleEnum_dataCreator,
					Person: message.Person{
						PersonGivenNames:  "Kat",
						PersonFamilyNames: "",
					},
				},
				{
					Role: message.PersonRoleEnum_publisher,
					Person: message.Person{
						PersonGivenNames:  "Joan",
						PersonFamilyNames: "",
					},
				},
			},
			ObjectIdentifier: []message.Identifier{
				{
					IdentifierType:  message.IdentifierTypeEnum_DOI,
					IdentifierValue: "10.5072/FK2/QAWS8O",
				},
			},
		},
		expected: []string{"objects/", "Kat", "10.5072/FK2/QAWS8O", "Joan", "ObjectTitle1", "artDesignItem"},
	},
	"No given name": {
		researchObject: message.ResearchObject{
			ObjectTitle: "ObjectTitle2",
			ObjectPersonRole: []message.PersonRole{
				{
					Role: message.PersonRoleEnum_dataCreator,
					Person: message.Person{
						PersonGivenNames:  "",
						PersonFamilyNames: "Winter",
					},
				},
				{
					Role: message.PersonRoleEnum_publisher,
					Person: message.Person{
						PersonGivenNames:  "",
						PersonFamilyNames: "Watson",
					},
				},
			},
			ObjectIdentifier: []message.Identifier{
				{
					IdentifierType:  message.IdentifierTypeEnum_DOI,
					IdentifierValue: "10.5072/FK2/QAWS81",
				},
			},
		},
		expected: []string{"objects/", "Winter", "10.5072/FK2/QAWS81", "Watson", "ObjectTitle2", "artDesignItem"},
	},
	"Both names": {
		researchObject: message.ResearchObject{
			ObjectTitle: "ObjectTitle3",
			ObjectPersonRole: []message.PersonRole{
				{
					Role: message.PersonRoleEnum_dataCreator,
					Person: message.Person{
						PersonGivenNames:  "Kat",
						PersonFamilyNames: "Winter",
					},
				},
				{
					Role: message.PersonRoleEnum_publisher,
					Person: message.Person{
						PersonGivenNames:  "Joan",
						PersonFamilyNames: "Watson",
					},
				},
			},
			ObjectIdentifier: []message.Identifier{
				{
					IdentifierType:  message.IdentifierTypeEnum_DOI,
					IdentifierValue: "10.5072/FK2/QAWS82",
				},
			},
		},
		expected: []string{"objects/", "Winter, Kat", "10.5072/FK2/QAWS82", "Watson, Joan", "ObjectTitle3", "artDesignItem"},
	},
	"No matching enums": {
		researchObject: message.ResearchObject{
			ObjectTitle: "ObjectTitle4",
			ObjectPersonRole: []message.PersonRole{
				{
					Role: message.PersonRoleEnum_editor,
					Person: message.Person{
						PersonGivenNames:  "Kat",
						PersonFamilyNames: "Winter",
					},
				},
				{
					Role: message.PersonRoleEnum_other,
					Person: message.Person{
						PersonGivenNames:  "Joan",
						PersonFamilyNames: "Watson",
					},
				},
			},
			ObjectIdentifier: []message.Identifier{
				{
					IdentifierType:  message.IdentifierTypeEnum_DOI,
					IdentifierValue: "10.5072/FK2/QAWS83",
				},
			},
		},
		expected: []string{"objects/", "10.5072/FK2/QAWS83", "ObjectTitle4", "artDesignItem"},
	},
}

// TestMetadataGeneration inspects some of the internal functionality of the
// channel adapter and ensures that from a given metadata set the correct fields
// are mapped into an object which then returns an expected CSV configuration
// for Archivematica. The test round-trips the CSV generation and checks that
// when it is consumed the row data is as is expected.
func TestMetadataGeneration(t *testing.T) {
	for name, tc := range metadataGenerationTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fs := afero.Afero{Fs: afero.NewBasePathFs(afero.NewMemMapFs(), "/")}
			transferSession := amclient.TransferSession{
				Metadata: amclient.NewMetadataSet(fs),
			}

			describeDataset(&transferSession, &tc.researchObject)
			transferSession.Metadata.Write()

			const mdFile string = "/metadata/metadata.csv"
			csvFile, err := fs.Open(mdFile)
			assert.NoError(t, err, "cannot open CSV file")

			csvObject := csv.NewReader(csvFile)

			// Read the first line so we can move onto the row data.
			_, err = csvObject.Read()
			assert.NoError(t, err, "cannot read CSV file")

			// Read the second line, the row data.
			record, err := csvObject.Read()
			assert.NoError(t, err, "cannot read CSV file")
			assert.Equal(t, len(record), len(tc.expected), "unexpected length in the CSV record")
			assert.EqualValues(t, tc.expected, record, "incorrect value in CSV record")
		})
	}
}
