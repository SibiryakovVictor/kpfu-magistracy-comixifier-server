package schema

import "fmt"

type ErrorResponse struct {
	Code int         `json:"code"`
	Msg  interface{} `json:"msg"`
}

func (r *ErrorResponse) MsgString() (string, error) {
	switch r.Msg.(type) {
	case string:
		return r.Msg.(string), nil
	case map[string]interface{}:
		mapMsg := r.Msg.(map[string]interface{})
		for k, v := range mapMsg {
			switch v.(type) {
			case []interface{}:
				first := v.([]interface{})[0]
				return fmt.Sprintf("%s: %s", k, first.(string)), nil
			default:
				return "", fmt.Errorf("unknown type of msg value: %T", v)
			}
		}

		return "", fmt.Errorf("msg is empty map")
	}

	return "", fmt.Errorf("unknown type of msg: %T", r.Msg)
}

type Response struct {
	Code   int    `json:"code"`
	CsCode int    `json:"cscode"`
	Ip     string `json:"ip"`
}

type UploadImageResponse struct {
	Response
	Data UploadImageResponseData `json:"data"`
}

type UploadImageResponseData struct {
	Uid       string `json:"uid"`
	Name      string `json:"name"`
	Thumbnail string `json:"thumbnail"`
	W         int    `json:"w"`
	H         int    `json:"h"`
	Filesize  int    `json:"int"`
}

type TransformResponse struct {
	Response
	Data TransformResponseData `json:"data"`
}

type TransformResponseData struct {
	TransId string `json:"trans_id"`
	Status  string `json:"status"`
}

type ProgressResponse struct {
	Response
	Data ProgressResponseData `json:"data"`
}

type ProgressResponseData struct {
	Status   string `json:"status"`
	Filesize uint32 `json:"filesize"`
}
