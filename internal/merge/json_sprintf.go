package merge

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

var ErrInvalidJSON = errors.New("invalid json")

type Merger interface {
	GetMerged(data []byte) ([]byte, error)
}

type JSONSprintfMerger struct {
	tpl string
}

func NewJSONSprintfMerger(tpl string) *JSONSprintfMerger {
	return &JSONSprintfMerger{
		tpl: tpl,
	}
}

func (s *JSONSprintfMerger) GetMerged(data []byte) ([]byte, error) {
	if !json.Valid(data) {
		return nil, ErrInvalidJSON
	}

	b := &bytes.Buffer{}
	if err := json.Indent(b, []byte(fmt.Sprintf(s.tpl, data)), "", " "); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
