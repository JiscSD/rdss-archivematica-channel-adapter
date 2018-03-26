package message

import "fmt"

type VocabularyReadRequest struct {
	VocabularyId int `json:"vocabularyId"`
}

func (m Message) VocabularyReadRequest() (*VocabularyReadRequest, error) {
	b, ok := m.MessageBody.(*VocabularyReadRequest)
	if !ok {
		return nil, fmt.Errorf("VocabularyReadRequest(): interface conversion error")
	}
	return b, nil
}

type VocabularyReadResponse struct {
	VocabularyId    int      `json:"vocabularyId"`
	VocabularyTerms []string `json:"vocabularyTerms"`
}

func (m Message) VocabularyReadResponse() (*VocabularyReadResponse, error) {
	b, ok := m.MessageBody.(*VocabularyReadResponse)
	if !ok {
		return nil, fmt.Errorf("VocabularyReadResponse(): interface conversion error")
	}
	return b, nil
}

type VocabularyPatchRequest struct {
	VocabularyId    int      `json:"vocabularyId"`
	VocabularyName  string   `json:"vocabularyName"`
	VocabularyTerms []string `json:"vocabularyTerms"`
}

func (m Message) VocabularyPatchRequest() (*VocabularyPatchRequest, error) {
	b, ok := m.MessageBody.(*VocabularyPatchRequest)
	if !ok {
		return nil, fmt.Errorf("VocabularyPatchRequest(): interface conversion error")
	}
	return b, nil
}
