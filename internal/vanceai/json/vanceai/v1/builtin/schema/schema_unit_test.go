package schema

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestErrorResponse_MsgString_Unit(t *testing.T) {
	type testCase struct {
		name    string
		jsonArg string
		wantErr error
		wantMsg string
	}
	tests := []testCase{
		{
			"correct: string",
			`{"msg": "msg content"}`,
			nil,
			"msg content",
		},
		{
			"correct: object",
			`{"msg": {"api_token": ["something wrong with api_token"]}}`,
			nil,
			"api_token: something wrong with api_token",
		},
		{
			"error: type of msg",
			`{"msg": 123}`,
			fmt.Errorf("unknown type of msg: float64"),
			"",
		},
		{
			"error: msg is empty map",
			`{"msg": {}}`,
			fmt.Errorf("msg is empty map"),
			"",
		},
		{
			"error: unknown type of msg value",
			`{"msg": {"api_token": 123}}`,
			fmt.Errorf("unknown type of msg value: float64"),
			"",
		},
	}

	for _, test := range tests {
		func(test testCase) {
			t.Run(test.name, func(t *testing.T) {
				errResp := &ErrorResponse{}
				decodeErr := json.Unmarshal([]byte(test.jsonArg), errResp)
				if decodeErr != nil {
					t.Logf("unexpected error of decoding json: %s", decodeErr.Error())
					t.FailNow()
				}

				msg, err := errResp.MsgString()
				if err != nil || test.wantErr != nil {
					isOnlyErrNil := err == nil && test.wantErr != nil
					isOnlyWantErrNil := err != nil && test.wantErr == nil
					isDifferentErrors := (err != nil && test.wantErr != nil) && (err.Error() != test.wantErr.Error())
					if isOnlyErrNil || isOnlyWantErrNil || isDifferentErrors {
						t.Logf(
							"MsgString() got error: %#v; expected error: %#v",
							err, test.wantErr,
						)
						t.FailNow()
					}
				}
				if msg != test.wantMsg {
					t.Logf("MsgString() got msg: %v, expected msg: %v", msg, test.wantMsg)
					t.FailNow()
				}
			})
		}(test)
	}
}
