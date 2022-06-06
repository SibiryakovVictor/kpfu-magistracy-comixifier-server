package errors

import "fmt"

type ApiError struct {
	code Code
	msg  string
}

func NewApiError(code Code, msg string) *ApiError {
	return &ApiError{
		code: code,
		msg:  msg,
	}
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("api error, code=%d; message=%s", e.code, e.msg)
}

type Code uint16

const (
	CodeIllegalParameter    Code = 10001
	CodeInternalError       Code = 10010
	CodeFileNotAvailable    Code = 10011
	CodeFileExceedsLimits   Code = 10012
	CodeJConfigParseError   Code = 10013
	CodeJobUnexpectedFailed Code = 10014

	CodeInvalidAPIToken     Code = 30014
	CodeInsufficientCredits Code = 30004
)

var codes = []Code{
	CodeIllegalParameter, CodeInternalError, CodeFileNotAvailable, CodeFileExceedsLimits,
	CodeJConfigParseError, CodeJobUnexpectedFailed, CodeInvalidAPIToken, CodeInsufficientCredits,
}

func MapCode(c int) (Code, error) {
	for _, code := range codes {
		if Code(c) == code {
			return code, nil
		}
	}

	return Code(0), fmt.Errorf("unknown code: %d", c)
}
