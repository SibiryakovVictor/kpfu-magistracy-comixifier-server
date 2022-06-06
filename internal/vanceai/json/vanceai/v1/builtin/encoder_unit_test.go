package builtin

import (
	"comixifier/internal/vanceai/json/vanceai/v1/jconfig"
	"encoding/json"
	"io"
	"log"
	"reflect"
	"testing"
)

func TestJConfigEncoder_Do_Unit(t *testing.T) {
	type testCase struct {
		name        string
		initJConfig *jconfig.Workflow
		wantJConfig func() *jconfig.Workflow
		wantErr     error
	}
	tests := []testCase{
		{
			name: "correct for workflow",
			initJConfig: jconfig.NewWorkflow(
				[]jconfig.Feature{
					&jconfig.ToongineerCartoonizer{
						Name_: "",
						Config_: &jconfig.ToongineerCartoonizerConfig{
							Module_:       "",
							ModuleParams_: &jconfig.ModuleParamsDefault{ModelName_: ""},
						},
					},
					&jconfig.ToongineerCartoonizer{
						Name_: "",
						Config_: &jconfig.ToongineerCartoonizerConfig{
							Module_:       "",
							ModuleParams_: &jconfig.ModuleParamsDefault{ModelName_: ""},
							OutParams_:    &jconfig.OutParams{SomeField: "azaza"},
						},
					},
				},
			),
			wantJConfig: func() *jconfig.Workflow {
				w := jconfig.NewWorkflow(
					[]jconfig.Feature{
						jconfig.NewToongineerCartoonizer(),
						jconfig.NewToongineerCartoonizer(),
					},
				)

				w.Features[len(w.Features)-1].Config().SetOutParams(&jconfig.OutParams{SomeField: "azaza"})

				return w
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wantJConfig := test.wantJConfig()
			encoder := NewJConfigEncoder()

			encJConfig, err := encoder.DoWorkflow(wantJConfig)
			if err != nil {
				t.Logf("encode jConfig: %s", err.Error())
				t.FailNow()
			}

			jConfigBytes, _ := io.ReadAll(encJConfig)
			log.Printf("encoded jConfig: %s", string(jConfigBytes))

			var mapJConfig map[string]interface{}
			json.Unmarshal(jConfigBytes, &mapJConfig)
			for k := range mapJConfig["config"].([]interface{}) {
				((mapJConfig["config"].([]interface{}))[k].(map[string]interface{}))["job"] =
					((mapJConfig["config"].([]interface{}))[k].(map[string]interface{}))["name"]
				delete((mapJConfig["config"].([]interface{}))[k].(map[string]interface{}), "name")
			}

			jConfigBytes, _ = json.Marshal(&mapJConfig)

			err = json.Unmarshal(jConfigBytes, test.initJConfig)
			if err != nil {
				t.Logf("decode encoded jConfig: %s", err.Error())
				t.FailNow()
			}

			if !reflect.DeepEqual(test.initJConfig, wantJConfig) {
				t.Logf("got jConfig: %#v; expected jConfig: %#v", test.initJConfig, wantJConfig)
				t.FailNow()
			}
		})
	}
}
