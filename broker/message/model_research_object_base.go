package message

import (
	"encoding/json"
	"errors"
	"reflect"
)

// ResearchObjectBase knows how to marshal/unmarshal the different subtypes
// allowed, e.g. an article is also a research object. Use the
// InferResearchObject method when you want to access the latter.
//
// Is there a better way to do this? E.g. ResearchObjectBase could be embedded
// instead.
type ResearchObjectBase struct {
	*Article
	*Dataset
	*ThesisDissertation
	*ResearchObject
}

func (m ResearchObjectBase) instance() (v interface{}) {
	if m.Article != nil {
		v = m.Article
	} else if m.Dataset != nil {
		v = m.Dataset
	} else if m.ThesisDissertation != nil {
		v = m.ThesisDissertation
	} else {
		v = m.ResearchObject
	}
	return
}

func (m ResearchObjectBase) MarshalJSON() ([]byte, error) {
	v := m.instance()
	if v == nil {
		return nil, errors.New("unitialized object")
	}
	return json.Marshal(v)
}

func (m *ResearchObjectBase) UnmarshalJSON(data []byte) error {
	proxy := struct {
		Kind string `json:"objectResourceType"`
	}{}
	if err := json.Unmarshal(data, &proxy); err != nil {
		return err
	}
	rt, ok := _ResourceTypeEnumNameToValue[proxy.Kind]
	if !ok {
		rt = ResourceTypeEnum_unknown
	}
	var err error
	switch rt {
	case ResourceTypeEnum_article:
		m.Article = &Article{}
		err = json.Unmarshal(data, m.Article)
	case ResourceTypeEnum_dataset:
		m.Dataset = &Dataset{}
		err = json.Unmarshal(data, m.Dataset)
	case ResourceTypeEnum_thesisDissertation:
		m.ThesisDissertation = &ThesisDissertation{}
		err = json.Unmarshal(data, m.ThesisDissertation)
	default:
		m.ResearchObject = &ResearchObject{}
		err = json.Unmarshal(data, m.ResearchObject)
	}
	if err != nil {
		return err
	}
	return nil
}

// fields that at shared between research object and its subclasses.
var fields = []string{
	"ObjectUUID",
	"ObjectTitle",
	"ObjectPersonRole",
	"ObjectDescription",
	"ObjectRights",
	"ObjectDate",
	"ObjectKeyword",
	"ObjectCategory",
	"ObjectResourceType",
	"ObjectValue",
	"ObjectIdentifier",
	"ObjectRelatedIdentifier",
	"ObjectOrganisationRole",
	"ObjectFile",
}

func (m ResearchObjectBase) InferResearchObject() *ResearchObject {
	instance := m.instance()
	instance_E := reflect.ValueOf(instance).Elem()
	researchObject := &ResearchObject{}
	researchObject_E := reflect.ValueOf(researchObject).Elem()

	for _, field := range fields {
		if !instance_E.IsValid() {
			continue
		}
		src_V := instance_E.FieldByName(field)
		dst_V := researchObject_E.FieldByName(field)

		dst_V.Set(src_V)
	}

	return researchObject
}
