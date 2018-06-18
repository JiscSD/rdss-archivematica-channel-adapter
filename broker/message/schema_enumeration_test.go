package message

import "testing"

func TestAccessTypeEnum_String(t *testing.T) {
	tests := []struct {
		t    AccessTypeEnum
		want string
	}{
		{AccessTypeEnum_open, "open"},
		{AccessTypeEnum_safeguarded, "safeguarded"},
		{AccessTypeEnum_controlled, "controlled"},
		{AccessTypeEnum_restricted, "restricted"},
		{AccessTypeEnum_closed, "closed"},
		{0, ""},
	}
	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("AccessTypeEnum.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestChecksumTypeEnum_String(t *testing.T) {
	tests := []struct {
		t    ChecksumTypeEnum
		want string
	}{
		{ChecksumTypeEnum_md5, "md5"},
		{ChecksumTypeEnum_sha256, "sha256"},
		{0, ""},
	}
	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("ChecksumTypeEnum.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestDateTypeEnum_String(t *testing.T) {
	tests := []struct {
		t    DateTypeEnum
		want string
	}{
		{DateTypeEnum_accepted, "accepted"},
		{DateTypeEnum_approved, "approved"},
		{DateTypeEnum_available, "available"},
		{DateTypeEnum_copyrighted, "copyrighted"},
		{DateTypeEnum_collected, "collected"},
		{DateTypeEnum_created, "created"},
		{DateTypeEnum_issued, "issued"},
		{DateTypeEnum_modified, "modified"},
		{DateTypeEnum_posted, "posted"},
		{DateTypeEnum_published, "published"},
		{0, ""},
	}
	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("DateTypeEnum.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestFileUseEnum_String(t *testing.T) {
	tests := []struct {
		t    FileUseEnum
		want string
	}{
		{FileUseEnum_originalFile, "originalFile"},
		{FileUseEnum_thumbnailImage, "thumbnailImage"},
		{FileUseEnum_extractedText, "extractedText"},
		{FileUseEnum_preservationMasterFile, "preservationMasterFile"},
		{FileUseEnum_intermediateFile, "intermediateFile"},
		{FileUseEnum_serviceFile, "serviceFile"},
		{FileUseEnum_transcript, "transcript"},
		{0, ""},
	}
	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("FileUseEnum.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestIdentifierTypeEnum_String(t *testing.T) {
	tests := []struct {
		t    IdentifierTypeEnum
		want string
	}{
		{IdentifierTypeEnum_ARK, "ARK"},
		{IdentifierTypeEnum_arXiv, "arXiv"},
		{IdentifierTypeEnum_bibcode, "bibcode"},
		{IdentifierTypeEnum_DOI, "DOI"},
		{IdentifierTypeEnum_EAN13, "EAN13"},
		{IdentifierTypeEnum_EISSN, "EISSN"},
		{IdentifierTypeEnum_Handle, "Handle"},
		{IdentifierTypeEnum_ISBN, "ISBN"},
		{IdentifierTypeEnum_ISSN, "ISSN"},
		{IdentifierTypeEnum_ISTC, "ISTC"},
		{IdentifierTypeEnum_LISSN, "LISSN"},
		{IdentifierTypeEnum_LSID, "LSID"},
		{IdentifierTypeEnum_PMID, "PMID"},
		{IdentifierTypeEnum_PUID, "PUID"},
		{IdentifierTypeEnum_PURL, "PURL"},
		{IdentifierTypeEnum_UPC, "UPC"},
		{IdentifierTypeEnum_URL, "URL"},
		{IdentifierTypeEnum_URN, "URN"},
		{0, ""},
	}
	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("IdentifierTypeEnum.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestObjectValueEnum_String(t *testing.T) {
	tests := []struct {
		t    ObjectValueEnum
		want string
	}{
		{ObjectValueEnum_normal, "normal"},
		{ObjectValueEnum_high, "high"},
		{ObjectValueEnum_veryHigh, "veryHigh"},
		{0, ""},
	}
	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("ObjectValueEnum.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestOrganisationRoleEnum_String(t *testing.T) {
	tests := []struct {
		t    OrganisationRoleEnum
		want string
	}{
		{OrganisationRoleEnum_funder, "funder"},
		{OrganisationRoleEnum_hostingInstitution, "hostingInstitution"},
		{OrganisationRoleEnum_rightsHolder, "rightsHolder"},
		{OrganisationRoleEnum_sponsor, "sponsor"},
		{OrganisationRoleEnum_publisher, "publisher"},
		{OrganisationRoleEnum_registrationAgency, "registrationAgency"},
		{OrganisationRoleEnum_registrationAuthority, "registrationAuthority"},
		{OrganisationRoleEnum_distributor, "distributor"},
		{OrganisationRoleEnum_advocacy, "advocacy"},
		{OrganisationRoleEnum_author, "author"},
		{0, ""},
	}
	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("OrganisationRoleEnum.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestOrganisationTypeEnum_String(t *testing.T) {
	tests := []struct {
		t    OrganisationTypeEnum
		want string
	}{
		{OrganisationTypeEnum_charity, "charity"},
		{OrganisationTypeEnum_commercial, "commercial"},
		{OrganisationTypeEnum_funder, "funder"},
		{OrganisationTypeEnum_furtherEducation, "furtherEducation"},
		{OrganisationTypeEnum_government, "government"},
		{OrganisationTypeEnum_health, "health"},
		{OrganisationTypeEnum_heritage, "heritage"},
		{OrganisationTypeEnum_higherEducation, "higherEducation"},
		{OrganisationTypeEnum_other, "other"},
		{OrganisationTypeEnum_professionalBody, "professionalBody"},
		{OrganisationTypeEnum_research, "research"},
		{OrganisationTypeEnum_school, "school"},
		{OrganisationTypeEnum_skills, "skills"},
		{OrganisationTypeEnum_billing, "billing"},
		{OrganisationTypeEnum_display, "display"},
		{0, ""},
	}
	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("OrganisationTypeEnum.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestPersonIdentifierTypeEnum_String(t *testing.T) {
	tests := []struct {
		t    PersonIdentifierTypeEnum
		want string
	}{
		{PersonIdentifierTypeEnum_ORCID, "ORCID"},
		{PersonIdentifierTypeEnum_researcherId, "researcherId"},
		{0, ""},
	}
	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("PersonIdentifierTypeEnum.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestPersonRoleEnum_String(t *testing.T) {
	tests := []struct {
		t    PersonRoleEnum
		want string
	}{
		{PersonRoleEnum_administrator, "administrator"},
		{PersonRoleEnum_contactPerson, "contactPerson"},
		{PersonRoleEnum_dataAnalyser, "dataAnalyser"},
		{PersonRoleEnum_dataCollector, "dataCollector"},
		{PersonRoleEnum_dataCreator, "dataCreator"},
		{PersonRoleEnum_dataManager, "dataManager"},
		{PersonRoleEnum_editor, "editor"},
		{PersonRoleEnum_investigator, "investigator"},
		{PersonRoleEnum_producer, "producer"},
		{PersonRoleEnum_projectLeader, "projectLeader"},
		{PersonRoleEnum_publisher, "publisher"},
		{PersonRoleEnum_projectMember, "projectMember"},
		{PersonRoleEnum_relatedPerson, "relatedPerson"},
		{PersonRoleEnum_researcher, "researcher"},
		{PersonRoleEnum_researcherGroup, "researcherGroup"},
		{PersonRoleEnum_rightsHolder, "rightsHolder"},
		{PersonRoleEnum_sponsor, "sponsor"},
		{PersonRoleEnum_supervisor, "supervisor"},
		{PersonRoleEnum_other, "other"},
		{PersonRoleEnum_author, "author"},
		{PersonRoleEnum_depositingUser, "depositingUser"},
		{0, ""},
	}
	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("PersonRoleEnum.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestPreservationEventTypeEnum_String(t *testing.T) {
	tests := []struct {
		t    PreservationEventTypeEnum
		want string
	}{
		{PreservationEventTypeEnum_capture, "capture"},
		{PreservationEventTypeEnum_compression, "compression"},
		{PreservationEventTypeEnum_creation, "creation"},
		{PreservationEventTypeEnum_deaccession, "deaccession"},
		{PreservationEventTypeEnum_decompression, "decompression"},
		{PreservationEventTypeEnum_decryption, "decryption"},
		{PreservationEventTypeEnum_deletion, "deletion"},
		{PreservationEventTypeEnum_digitalSignatureValidation, "digitalSignatureValidation"},
		{PreservationEventTypeEnum_download, "download"},
		{PreservationEventTypeEnum_fixityCheck, "fixityCheck"},
		{PreservationEventTypeEnum_ingestion, "ingestion"},
		{PreservationEventTypeEnum_messageDigestCalculation, "messageDigestCalculation"},
		{PreservationEventTypeEnum_migration, "migration"},
		{PreservationEventTypeEnum_normalisation, "normalisation"},
		{PreservationEventTypeEnum_replication, "replication"},
		{PreservationEventTypeEnum_update, "update"},
		{PreservationEventTypeEnum_validation, "validation"},
		{PreservationEventTypeEnum_virusCheck, "virusCheck"},
		{0, ""},
	}
	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("PreservationEventTypeEnum.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestRelationTypeEnum_String(t *testing.T) {
	tests := []struct {
		t    RelationTypeEnum
		want string
	}{
		{RelationTypeEnum_cites, "cites"},
		{RelationTypeEnum_isCitedBy, "isCitedBy"},
		{RelationTypeEnum_isSupplementTo, "isSupplementTo"},
		{RelationTypeEnum_isSupplementedBy, "isSupplementedBy"},
		{RelationTypeEnum_continues, "continues"},
		{RelationTypeEnum_isContinuedBy, "isContinuedBy"},
		{RelationTypeEnum_hasMetadata, "hasMetadata"},
		{RelationTypeEnum_isMetadataFor, "isMetadataFor"},
		{RelationTypeEnum_isNewVersionOf, "isNewVersionOf"},
		{RelationTypeEnum_isPreviousVersionOf, "isPreviousVersionOf"},
		{RelationTypeEnum_isPartOf, "isPartOf"},
		{RelationTypeEnum_hasPart, "hasPart"},
		{RelationTypeEnum_isReferencedBy, "isReferencedBy"},
		{RelationTypeEnum_references, "references"},
		{RelationTypeEnum_isDocumentedBy, "isDocumentedBy"},
		{RelationTypeEnum_documents, "documents"},
		{RelationTypeEnum_isCompiledBy, "isCompiledBy"},
		{RelationTypeEnum_compiles, "compiles"},
		{RelationTypeEnum_isVariantFormOf, "isVariantFormOf"},
		{RelationTypeEnum_isOriginalFormOf, "isOriginalFormOf"},
		{RelationTypeEnum_isIdenticalTo, "isIdenticalTo"},
		{RelationTypeEnum_isReviewedBy, "isReviewedBy"},
		{RelationTypeEnum_reviews, "reviews"},
		{RelationTypeEnum_isDerivedFrom, "isDerivedFrom"},
		{RelationTypeEnum_isSourceOf, "isSourceOf"},
		{RelationTypeEnum_isCommentOn, "isCommentOn"},
		{RelationTypeEnum_hasComment, "hasComment"},
		{RelationTypeEnum_isReplyTo, "isReplyTo"},
		{RelationTypeEnum_hasReply, "hasReply"},
		{RelationTypeEnum_basedOnData, "basedOnData"},
		{RelationTypeEnum_hasRelatedMaterial, "hasRelatedMaterial"},
		{RelationTypeEnum_isBasedOn, "isBasedOn"},
		{RelationTypeEnum_isBasisFor, "isBasisFor"},
		{RelationTypeEnum_requires, "requires"},
		{RelationTypeEnum_isRequiredBy, "isRequiredBy"},
		{RelationTypeEnum_hasParent, "hasParent"},
		{RelationTypeEnum_isParentOf, "isParentOf"},
		{0, ""},
	}
	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("RelationTypeEnum.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestResourceTypeEnum_String(t *testing.T) {
	tests := []struct {
		t    ResourceTypeEnum
		want string
	}{
		{ResourceTypeEnum_artDesignItem, "artDesignItem"},
		{ResourceTypeEnum_article, "article"},
		{ResourceTypeEnum_audio, "audio"},
		{ResourceTypeEnum_book, "book"},
		{ResourceTypeEnum_bookSection, "bookSection"},
		{ResourceTypeEnum_conferenceWorkshopItem, "conferenceWorkshopItem"},
		{ResourceTypeEnum_dataset, "dataset"},
		{ResourceTypeEnum_examPaper, "examPaper"},
		{ResourceTypeEnum_image, "image"},
		{ResourceTypeEnum_journal, "journal"},
		{ResourceTypeEnum_learningObject, "learningObject"},
		{ResourceTypeEnum_movingImage, "movingImage"},
		{ResourceTypeEnum_musicComposition, "musicComposition"},
		{ResourceTypeEnum_other, "other"},
		{ResourceTypeEnum_patent, "patent"},
		{ResourceTypeEnum_performance, "performance"},
		{ResourceTypeEnum_preprint, "preprint"},
		{ResourceTypeEnum_report, "report"},
		{ResourceTypeEnum_review, "review"},
		{ResourceTypeEnum_showExhibition, "showExhibition"},
		{ResourceTypeEnum_software, "software"},
		{ResourceTypeEnum_text, "text"},
		{ResourceTypeEnum_thesisDissertation, "thesisDissertation"},
		{ResourceTypeEnum_unknown, "unknown"},
		{ResourceTypeEnum_website, "website"},
		{ResourceTypeEnum_workflow, "workflow"},
		{ResourceTypeEnum_equipment, "equipment"},
		{0, ""},
	}
	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("ResourceTypeEnum.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestStorageStatusEnum_String(t *testing.T) {
	tests := []struct {
		t    StorageStatusEnum
		want string
	}{
		{StorageStatusEnum_online, "online"},
		{StorageStatusEnum_nearline, "nearline"},
		{StorageStatusEnum_offline, "offline"},
		{0, ""},
	}
	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("StorageStatusEnum.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestStorageTypeEnum_String(t *testing.T) {
	tests := []struct {
		t    StorageTypeEnum
		want string
	}{
		{StorageTypeEnum_S3, "S3"},
		{StorageTypeEnum_HTTP, "HTTP"},
		{0, ""},
	}
	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("StorageTypeEnum.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestUploadStatusEnum_String(t *testing.T) {
	tests := []struct {
		t    UploadStatusEnum
		want string
	}{
		{UploadStatusEnum_uploadStarted, "uploadStarted"},
		{UploadStatusEnum_uploadComplete, "uploadComplete"},
		{UploadStatusEnum_uploadAborted, "uploadAborted"},
		{0, ""},
	}
	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("UploadStatusEnum.String() = %v, want %v", got, tt.want)
		}
	}
}
