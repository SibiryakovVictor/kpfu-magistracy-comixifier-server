package builtin

import (
	"bytes"
	"comixifier/internal/vanceai/json/vanceai/v1/jconfig"
	"encoding/json"
	"fmt"
	"io"
)

type JConfigEncoder struct {
}

func NewJConfigEncoder() *JConfigEncoder {
	return &JConfigEncoder{}
}

func (e *JConfigEncoder) DoSingleJob(j *jconfig.SingleJob) (io.Reader, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(j)
	if err != nil {
		return nil, fmt.Errorf("encode jConfig to json: %w", err)
	}
	return buf, nil
}

func (e *JConfigEncoder) DoWorkflow(w *jconfig.Workflow) (io.Reader, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(w)
	if err != nil {
		return nil, fmt.Errorf("encode jConfig to json: %w", err)
	}

	var mapJConfig map[string]interface{}
	err = json.NewDecoder(buf).Decode(&mapJConfig)
	if err != nil {
		return nil, fmt.Errorf("decode jConfig to map: %w", err)
	}

	for k := range mapJConfig["config"].([]interface{}) {
		((mapJConfig["config"].([]interface{}))[k].(map[string]interface{}))["name"] =
			((mapJConfig["config"].([]interface{}))[k].(map[string]interface{}))["job"]

		delete((mapJConfig["config"].([]interface{}))[k].(map[string]interface{}), "job")
	}

	buf.Reset()
	err = json.NewEncoder(buf).Encode(mapJConfig)
	if err != nil {
		return nil, fmt.Errorf("encode jConfig to json: %w", err)
	}

	return buf, nil
}
