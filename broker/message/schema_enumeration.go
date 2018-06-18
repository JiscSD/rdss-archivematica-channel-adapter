package message

// https://www.jisc.ac.uk/rdss/schema/research_object.json/#/definitions/access (.accessType, :required)
type AccessTypeEnum int

const (
	_                                  = iota
	AccessTypeEnum_open AccessTypeEnum = iota
	AccessTypeEnum_safeguarded
	AccessTypeEnum_controlled
	AccessTypeEnum_restricted
	AccessTypeEnum_closed
)

func (t AccessTypeEnum) String() string {
	switch t {
	case AccessTypeEnum_open:
		return "open"
	case AccessTypeEnum_safeguarded:
		return "safeguarded"
	case AccessTypeEnum_controlled:
		return "controlled"
	case AccessTypeEnum_restricted:
		return "restricted"
	case AccessTypeEnum_closed:
		return "closed"
	}
	return ""
}

// https://www.jisc.ac.uk/rdss/schema/intellectual_asset.json/#/definitions/checksum (.checksumType, :required)
type ChecksumTypeEnum int

const (
	_                                     = iota
	ChecksumTypeEnum_md5 ChecksumTypeEnum = iota
	ChecksumTypeEnum_sha256
)

func (t ChecksumTypeEnum) String() string {
	switch t {
	case ChecksumTypeEnum_md5:
		return "md5"
	case ChecksumTypeEnum_sha256:
		return "sha256"
	}
	return ""
}

// https://www.jisc.ac.uk/rdss/schema/research_object.json/#/definitions/date (.dateType, :required)
type DateTypeEnum int

const (
	_                                  = iota
	DateTypeEnum_accepted DateTypeEnum = iota
	DateTypeEnum_approved
	DateTypeEnum_available
	DateTypeEnum_copyrighted
	DateTypeEnum_collected
	DateTypeEnum_created
	DateTypeEnum_issued
	DateTypeEnum_modified
	DateTypeEnum_posted
	DateTypeEnum_published
)

func (t DateTypeEnum) String() string {
	switch t {
	case DateTypeEnum_accepted:
		return "accepted"
	case DateTypeEnum_approved:
		return "approved"
	case DateTypeEnum_available:
		return "available"
	case DateTypeEnum_copyrighted:
		return "copyrighted"
	case DateTypeEnum_collected:
		return "collected"
	case DateTypeEnum_created:
		return "created"
	case DateTypeEnum_issued:
		return "issued"
	case DateTypeEnum_modified:
		return "modified"
	case DateTypeEnum_posted:
		return "posted"
	case DateTypeEnum_published:
		return "published"
	}
	return ""
}

// https://www.jisc.ac.uk/rdss/schema/intellectual_asset.json/#/definitions/file (.fileUse, :required)
type FileUseEnum int

const (
	_                                    = iota
	FileUseEnum_originalFile FileUseEnum = iota
	FileUseEnum_thumbnailImage
	FileUseEnum_extractedText
	FileUseEnum_preservationMasterFile
	FileUseEnum_intermediateFile
	FileUseEnum_serviceFile
	FileUseEnum_transcript
)

func (t FileUseEnum) String() string {
	switch t {
	case FileUseEnum_originalFile:
		return "originalFile"
	case FileUseEnum_thumbnailImage:
		return "thumbnailImage"
	case FileUseEnum_extractedText:
		return "extractedText"
	case FileUseEnum_preservationMasterFile:
		return "preservationMasterFile"
	case FileUseEnum_intermediateFile:
		return "intermediateFile"
	case FileUseEnum_serviceFile:
		return "serviceFile"
	case FileUseEnum_transcript:
		return "transcript"
	}
	return ""
}

// https://www.jisc.ac.uk/rdss/schema/research_object.json/#/definitions/identifier (.identifierType, :required)
type IdentifierTypeEnum int

const (
	_                                         = iota
	IdentifierTypeEnum_ARK IdentifierTypeEnum = iota
	IdentifierTypeEnum_arXiv
	IdentifierTypeEnum_bibcode
	IdentifierTypeEnum_DOI
	IdentifierTypeEnum_EAN13
	IdentifierTypeEnum_EISSN
	IdentifierTypeEnum_Handle
	IdentifierTypeEnum_ISBN
	IdentifierTypeEnum_ISSN
	IdentifierTypeEnum_ISTC
	IdentifierTypeEnum_LISSN
	IdentifierTypeEnum_LSID
	IdentifierTypeEnum_PMID
	IdentifierTypeEnum_PUID
	IdentifierTypeEnum_PURL
	IdentifierTypeEnum_SourceID
	IdentifierTypeEnum_UPC
	IdentifierTypeEnum_URL
	IdentifierTypeEnum_URN
)

func (t IdentifierTypeEnum) String() string {
	switch t {
	case IdentifierTypeEnum_ARK:
		return "ARK"
	case IdentifierTypeEnum_arXiv:
		return "arXiv"
	case IdentifierTypeEnum_bibcode:
		return "bibcode"
	case IdentifierTypeEnum_DOI:
		return "DOI"
	case IdentifierTypeEnum_EAN13:
		return "EAN13"
	case IdentifierTypeEnum_EISSN:
		return "EISSN"
	case IdentifierTypeEnum_Handle:
		return "Handle"
	case IdentifierTypeEnum_ISBN:
		return "ISBN"
	case IdentifierTypeEnum_ISSN:
		return "ISSN"
	case IdentifierTypeEnum_ISTC:
		return "ISTC"
	case IdentifierTypeEnum_LISSN:
		return "LISSN"
	case IdentifierTypeEnum_LSID:
		return "LSID"
	case IdentifierTypeEnum_PMID:
		return "PMID"
	case IdentifierTypeEnum_PUID:
		return "PUID"
	case IdentifierTypeEnum_PURL:
		return "PURL"
	case IdentifierTypeEnum_SourceID:
		return "SourceID"
	case IdentifierTypeEnum_UPC:
		return "UPC"
	case IdentifierTypeEnum_URL:
		return "URL"
	case IdentifierTypeEnum_URN:
		return "URN"
	}
	return ""
}

// https://www.jisc.ac.uk/rdss/schema/research_object.json/#/definitions/object (.objectValue, :required)
type ObjectValueEnum int

const (
	_                                      = iota
	ObjectValueEnum_normal ObjectValueEnum = iota
	ObjectValueEnum_high
	ObjectValueEnum_veryHigh
)

func (t ObjectValueEnum) String() string {
	switch t {
	case ObjectValueEnum_normal:
		return "normal"
	case ObjectValueEnum_high:
		return "high"
	case ObjectValueEnum_veryHigh:
		return "veryHigh"
	}
	return ""
}

// https://www.jisc.ac.uk/rdss/schema/research_object.json/#/definitions/organisationRole (.role, :required)
type OrganisationRoleEnum int

const (
	_                                                = iota
	OrganisationRoleEnum_funder OrganisationRoleEnum = iota
	OrganisationRoleEnum_hostingInstitution
	OrganisationRoleEnum_rightsHolder
	OrganisationRoleEnum_sponsor
	OrganisationRoleEnum_publisher
	OrganisationRoleEnum_registrationAgency
	OrganisationRoleEnum_registrationAuthority
	OrganisationRoleEnum_distributor
	OrganisationRoleEnum_advocacy
	OrganisationRoleEnum_author
)

func (t OrganisationRoleEnum) String() string {
	switch t {
	case OrganisationRoleEnum_funder:
		return "funder"
	case OrganisationRoleEnum_hostingInstitution:
		return "hostingInstitution"
	case OrganisationRoleEnum_rightsHolder:
		return "rightsHolder"
	case OrganisationRoleEnum_sponsor:
		return "sponsor"
	case OrganisationRoleEnum_publisher:
		return "publisher"
	case OrganisationRoleEnum_registrationAgency:
		return "registrationAgency"
	case OrganisationRoleEnum_registrationAuthority:
		return "registrationAuthority"
	case OrganisationRoleEnum_distributor:
		return "distributor"
	case OrganisationRoleEnum_advocacy:
		return "advocacy"
	case OrganisationRoleEnum_author:
		return "author"
	}
	return ""
}

// https://www.jisc.ac.uk/rdss/schema/material_asset.json/#/definitions/organisation (.organisationType, :required)
type OrganisationTypeEnum int

const (
	_                                                 = iota
	OrganisationTypeEnum_charity OrganisationTypeEnum = iota
	OrganisationTypeEnum_commercial
	OrganisationTypeEnum_funder
	OrganisationTypeEnum_furtherEducation
	OrganisationTypeEnum_government
	OrganisationTypeEnum_health
	OrganisationTypeEnum_heritage
	OrganisationTypeEnum_higherEducation
	OrganisationTypeEnum_other
	OrganisationTypeEnum_professionalBody
	OrganisationTypeEnum_research
	OrganisationTypeEnum_school
	OrganisationTypeEnum_skills
	OrganisationTypeEnum_billing
	OrganisationTypeEnum_display
)

func (t OrganisationTypeEnum) String() string {
	switch t {
	case OrganisationTypeEnum_charity:
		return "charity"
	case OrganisationTypeEnum_commercial:
		return "commercial"
	case OrganisationTypeEnum_funder:
		return "funder"
	case OrganisationTypeEnum_furtherEducation:
		return "furtherEducation"
	case OrganisationTypeEnum_government:
		return "government"
	case OrganisationTypeEnum_health:
		return "health"
	case OrganisationTypeEnum_heritage:
		return "heritage"
	case OrganisationTypeEnum_higherEducation:
		return "higherEducation"
	case OrganisationTypeEnum_other:
		return "other"
	case OrganisationTypeEnum_professionalBody:
		return "professionalBody"
	case OrganisationTypeEnum_research:
		return "research"
	case OrganisationTypeEnum_school:
		return "school"
	case OrganisationTypeEnum_skills:
		return "skills"
	case OrganisationTypeEnum_billing:
		return "billing"
	case OrganisationTypeEnum_display:
		return "display"
	}
	return ""
}

// https://www.jisc.ac.uk/rdss/schema/material_asset.json/#/definitions/personIdentifier (.personIdentifierType, :required)
type PersonIdentifierTypeEnum int

const (
	_                                                       = iota
	PersonIdentifierTypeEnum_ORCID PersonIdentifierTypeEnum = iota
	PersonIdentifierTypeEnum_researcherId
)

func (t PersonIdentifierTypeEnum) String() string {
	switch t {
	case PersonIdentifierTypeEnum_ORCID:
		return "ORCID"
	case PersonIdentifierTypeEnum_researcherId:
		return "researcherId"
	}
	return ""
}

// https://www.jisc.ac.uk/rdss/schema/research_object.json/#/definitions/personRole (.role, :required)
type PersonRoleEnum int

const (
	_                                           = iota
	PersonRoleEnum_administrator PersonRoleEnum = iota
	PersonRoleEnum_contactPerson
	PersonRoleEnum_dataAnalyser
	PersonRoleEnum_dataCollector
	PersonRoleEnum_dataCreator
	PersonRoleEnum_dataManager
	PersonRoleEnum_editor
	PersonRoleEnum_investigator
	PersonRoleEnum_producer
	PersonRoleEnum_projectLeader
	PersonRoleEnum_publisher
	PersonRoleEnum_projectMember
	PersonRoleEnum_relatedPerson
	PersonRoleEnum_researcher
	PersonRoleEnum_researcherGroup
	PersonRoleEnum_rightsHolder
	PersonRoleEnum_sponsor
	PersonRoleEnum_supervisor
	PersonRoleEnum_other
	PersonRoleEnum_author
	PersonRoleEnum_depositingUser
)

func (t PersonRoleEnum) String() string {
	switch t {
	case PersonRoleEnum_administrator:
		return "administrator"
	case PersonRoleEnum_contactPerson:
		return "contactPerson"
	case PersonRoleEnum_dataAnalyser:
		return "dataAnalyser"
	case PersonRoleEnum_dataCollector:
		return "dataCollector"
	case PersonRoleEnum_dataCreator:
		return "dataCreator"
	case PersonRoleEnum_dataManager:
		return "dataManager"
	case PersonRoleEnum_editor:
		return "editor"
	case PersonRoleEnum_investigator:
		return "investigator"
	case PersonRoleEnum_producer:
		return "producer"
	case PersonRoleEnum_projectLeader:
		return "projectLeader"
	case PersonRoleEnum_publisher:
		return "publisher"
	case PersonRoleEnum_projectMember:
		return "projectMember"
	case PersonRoleEnum_relatedPerson:
		return "relatedPerson"
	case PersonRoleEnum_researcher:
		return "researcher"
	case PersonRoleEnum_researcherGroup:
		return "researcherGroup"
	case PersonRoleEnum_rightsHolder:
		return "rightsHolder"
	case PersonRoleEnum_sponsor:
		return "sponsor"
	case PersonRoleEnum_supervisor:
		return "supervisor"
	case PersonRoleEnum_other:
		return "other"
	case PersonRoleEnum_author:
		return "author"
	case PersonRoleEnum_depositingUser:
		return "depositingUser"
	}
	return ""
}

// https://www.jisc.ac.uk/rdss/schema/intellectual_asset.json/#/definitions/preservationEvent (.preservationEventType, :required)
type PreservationEventTypeEnum int

const (
	_                                                           = iota
	PreservationEventTypeEnum_capture PreservationEventTypeEnum = iota
	PreservationEventTypeEnum_compression
	PreservationEventTypeEnum_creation
	PreservationEventTypeEnum_deaccession
	PreservationEventTypeEnum_decompression
	PreservationEventTypeEnum_decryption
	PreservationEventTypeEnum_deletion
	PreservationEventTypeEnum_digitalSignatureValidation
	PreservationEventTypeEnum_download
	PreservationEventTypeEnum_fixityCheck
	PreservationEventTypeEnum_ingestion
	PreservationEventTypeEnum_messageDigestCalculation
	PreservationEventTypeEnum_migration
	PreservationEventTypeEnum_normalisation
	PreservationEventTypeEnum_replication
	PreservationEventTypeEnum_update
	PreservationEventTypeEnum_validation
	PreservationEventTypeEnum_virusCheck
)

func (t PreservationEventTypeEnum) String() string {
	switch t {
	case PreservationEventTypeEnum_capture:
		return "capture"
	case PreservationEventTypeEnum_compression:
		return "compression"
	case PreservationEventTypeEnum_creation:
		return "creation"
	case PreservationEventTypeEnum_deaccession:
		return "deaccession"
	case PreservationEventTypeEnum_decompression:
		return "decompression"
	case PreservationEventTypeEnum_decryption:
		return "decryption"
	case PreservationEventTypeEnum_deletion:
		return "deletion"
	case PreservationEventTypeEnum_digitalSignatureValidation:
		return "digitalSignatureValidation"
	case PreservationEventTypeEnum_download:
		return "download"
	case PreservationEventTypeEnum_fixityCheck:
		return "fixityCheck"
	case PreservationEventTypeEnum_ingestion:
		return "ingestion"
	case PreservationEventTypeEnum_messageDigestCalculation:
		return "messageDigestCalculation"
	case PreservationEventTypeEnum_migration:
		return "migration"
	case PreservationEventTypeEnum_normalisation:
		return "normalisation"
	case PreservationEventTypeEnum_replication:
		return "replication"
	case PreservationEventTypeEnum_update:
		return "update"
	case PreservationEventTypeEnum_validation:
		return "validation"
	case PreservationEventTypeEnum_virusCheck:
		return "virusCheck"
	}
	return ""
}

// https://www.jisc.ac.uk/rdss/schema/research_object.json/#/definitions/identifier (.relationType, :required)
type RelationTypeEnum int

const (
	_                                       = iota
	RelationTypeEnum_cites RelationTypeEnum = iota
	RelationTypeEnum_isCitedBy
	RelationTypeEnum_isSupplementTo
	RelationTypeEnum_isSupplementedBy
	RelationTypeEnum_continues
	RelationTypeEnum_isContinuedBy
	RelationTypeEnum_hasMetadata
	RelationTypeEnum_isMetadataFor
	RelationTypeEnum_isNewVersionOf
	RelationTypeEnum_isPreviousVersionOf
	RelationTypeEnum_isPartOf
	RelationTypeEnum_hasPart
	RelationTypeEnum_isReferencedBy
	RelationTypeEnum_references
	RelationTypeEnum_isDocumentedBy
	RelationTypeEnum_documents
	RelationTypeEnum_isCompiledBy
	RelationTypeEnum_compiles
	RelationTypeEnum_isVariantFormOf
	RelationTypeEnum_isOriginalFormOf
	RelationTypeEnum_isIdenticalTo
	RelationTypeEnum_isReviewedBy
	RelationTypeEnum_reviews
	RelationTypeEnum_isDerivedFrom
	RelationTypeEnum_isSourceOf
	RelationTypeEnum_isCommentOn
	RelationTypeEnum_hasComment
	RelationTypeEnum_isReplyTo
	RelationTypeEnum_hasReply
	RelationTypeEnum_basedOnData
	RelationTypeEnum_hasRelatedMaterial
	RelationTypeEnum_isBasedOn
	RelationTypeEnum_isBasisFor
	RelationTypeEnum_requires
	RelationTypeEnum_isRequiredBy
	RelationTypeEnum_hasParent
	RelationTypeEnum_isParentOf
)

func (t RelationTypeEnum) String() string {
	switch t {
	case RelationTypeEnum_cites:
		return "cites"
	case RelationTypeEnum_isCitedBy:
		return "isCitedBy"
	case RelationTypeEnum_isSupplementTo:
		return "isSupplementTo"
	case RelationTypeEnum_isSupplementedBy:
		return "isSupplementedBy"
	case RelationTypeEnum_continues:
		return "continues"
	case RelationTypeEnum_isContinuedBy:
		return "isContinuedBy"
	case RelationTypeEnum_hasMetadata:
		return "hasMetadata"
	case RelationTypeEnum_isMetadataFor:
		return "isMetadataFor"
	case RelationTypeEnum_isNewVersionOf:
		return "isNewVersionOf"
	case RelationTypeEnum_isPreviousVersionOf:
		return "isPreviousVersionOf"
	case RelationTypeEnum_isPartOf:
		return "isPartOf"
	case RelationTypeEnum_hasPart:
		return "hasPart"
	case RelationTypeEnum_isReferencedBy:
		return "isReferencedBy"
	case RelationTypeEnum_references:
		return "references"
	case RelationTypeEnum_isDocumentedBy:
		return "isDocumentedBy"
	case RelationTypeEnum_documents:
		return "documents"
	case RelationTypeEnum_isCompiledBy:
		return "isCompiledBy"
	case RelationTypeEnum_compiles:
		return "compiles"
	case RelationTypeEnum_isVariantFormOf:
		return "isVariantFormOf"
	case RelationTypeEnum_isOriginalFormOf:
		return "isOriginalFormOf"
	case RelationTypeEnum_isIdenticalTo:
		return "isIdenticalTo"
	case RelationTypeEnum_isReviewedBy:
		return "isReviewedBy"
	case RelationTypeEnum_reviews:
		return "reviews"
	case RelationTypeEnum_isDerivedFrom:
		return "isDerivedFrom"
	case RelationTypeEnum_isSourceOf:
		return "isSourceOf"
	case RelationTypeEnum_isCommentOn:
		return "isCommentOn"
	case RelationTypeEnum_hasComment:
		return "hasComment"
	case RelationTypeEnum_isReplyTo:
		return "isReplyTo"
	case RelationTypeEnum_hasReply:
		return "hasReply"
	case RelationTypeEnum_basedOnData:
		return "basedOnData"
	case RelationTypeEnum_hasRelatedMaterial:
		return "hasRelatedMaterial"
	case RelationTypeEnum_isBasedOn:
		return "isBasedOn"
	case RelationTypeEnum_isBasisFor:
		return "isBasisFor"
	case RelationTypeEnum_requires:
		return "requires"
	case RelationTypeEnum_isRequiredBy:
		return "isRequiredBy"
	case RelationTypeEnum_hasParent:
		return "hasParent"
	case RelationTypeEnum_isParentOf:
		return "isParentOf"
	}
	return ""
}

// https://www.jisc.ac.uk/rdss/schema/research_object.json/#/definitions/object (.objectResourceType, :required)
type ResourceTypeEnum int

const (
	_                                               = iota
	ResourceTypeEnum_artDesignItem ResourceTypeEnum = iota
	ResourceTypeEnum_article
	ResourceTypeEnum_audio
	ResourceTypeEnum_book
	ResourceTypeEnum_bookSection
	ResourceTypeEnum_conferenceWorkshopItem
	ResourceTypeEnum_dataset
	ResourceTypeEnum_examPaper
	ResourceTypeEnum_image
	ResourceTypeEnum_journal
	ResourceTypeEnum_learningObject
	ResourceTypeEnum_movingImage
	ResourceTypeEnum_musicComposition
	ResourceTypeEnum_other
	ResourceTypeEnum_patent
	ResourceTypeEnum_performance
	ResourceTypeEnum_preprint
	ResourceTypeEnum_report
	ResourceTypeEnum_review
	ResourceTypeEnum_showExhibition
	ResourceTypeEnum_software
	ResourceTypeEnum_text
	ResourceTypeEnum_thesisDissertation
	ResourceTypeEnum_unknown
	ResourceTypeEnum_website
	ResourceTypeEnum_workflow
	ResourceTypeEnum_equipment
)

func (t ResourceTypeEnum) String() string {
	switch t {
	case ResourceTypeEnum_artDesignItem:
		return "artDesignItem"
	case ResourceTypeEnum_article:
		return "article"
	case ResourceTypeEnum_audio:
		return "audio"
	case ResourceTypeEnum_book:
		return "book"
	case ResourceTypeEnum_bookSection:
		return "bookSection"
	case ResourceTypeEnum_conferenceWorkshopItem:
		return "conferenceWorkshopItem"
	case ResourceTypeEnum_dataset:
		return "dataset"
	case ResourceTypeEnum_examPaper:
		return "examPaper"
	case ResourceTypeEnum_image:
		return "image"
	case ResourceTypeEnum_journal:
		return "journal"
	case ResourceTypeEnum_learningObject:
		return "learningObject"
	case ResourceTypeEnum_movingImage:
		return "movingImage"
	case ResourceTypeEnum_musicComposition:
		return "musicComposition"
	case ResourceTypeEnum_other:
		return "other"
	case ResourceTypeEnum_patent:
		return "patent"
	case ResourceTypeEnum_performance:
		return "performance"
	case ResourceTypeEnum_preprint:
		return "preprint"
	case ResourceTypeEnum_report:
		return "report"
	case ResourceTypeEnum_review:
		return "review"
	case ResourceTypeEnum_showExhibition:
		return "showExhibition"
	case ResourceTypeEnum_software:
		return "software"
	case ResourceTypeEnum_text:
		return "text"
	case ResourceTypeEnum_thesisDissertation:
		return "thesisDissertation"
	case ResourceTypeEnum_unknown:
		return "unknown"
	case ResourceTypeEnum_website:
		return "website"
	case ResourceTypeEnum_workflow:
		return "workflow"
	case ResourceTypeEnum_equipment:
		return "equipment"
	}
	return ""
}

// https://www.jisc.ac.uk/rdss/schema/intellectual_asset.json/#/definitions/file (.fileStorageStatus, :required)
type StorageStatusEnum int

const (
	_                                          = iota
	StorageStatusEnum_online StorageStatusEnum = iota
	StorageStatusEnum_nearline
	StorageStatusEnum_offline
)

func (t StorageStatusEnum) String() string {
	switch t {
	case StorageStatusEnum_online:
		return "online"
	case StorageStatusEnum_nearline:
		return "nearline"
	case StorageStatusEnum_offline:
		return "offline"
	}
	return ""
}

// https://www.jisc.ac.uk/rdss/schema/intellectual_asset.json/#/definitions/file (.fileStorageType, :required)
type StorageTypeEnum int

const (
	_                                  = iota
	StorageTypeEnum_S3 StorageTypeEnum = iota
	StorageTypeEnum_HTTP
)

func (t StorageTypeEnum) String() string {
	switch t {
	case StorageTypeEnum_S3:
		return "S3"
	case StorageTypeEnum_HTTP:
		return "HTTP"
	}
	return ""
}

// https://www.jisc.ac.uk/rdss/schema/intellectual_asset.json/#/definitions/file (.fileUploadStatus, :required)
type UploadStatusEnum int

const (
	_                                               = iota
	UploadStatusEnum_uploadStarted UploadStatusEnum = iota
	UploadStatusEnum_uploadComplete
	UploadStatusEnum_uploadAborted
)

func (t UploadStatusEnum) String() string {
	switch t {
	case UploadStatusEnum_uploadStarted:
		return "uploadStarted"
	case UploadStatusEnum_uploadComplete:
		return "uploadComplete"
	case UploadStatusEnum_uploadAborted:
		return "uploadAborted"
	}
	return ""
}
