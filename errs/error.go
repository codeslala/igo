package errs

import (
	"encoding/json"
	"errors"

	"github.com/codeslala/igo/must"
)

type Error struct {
	Code Code   `json:"code"`
	Msg  string `json:"msg"`
}

func (e *Error) Error() string {
	if e.Msg != "" {
		return e.Code.String() + "[" + e.Msg + "]"
	}
	return e.Code.String()
}

func (e *Error) Json() string {
	return string(must.Byte(json.Marshal(e)))
}

// Decode tries to decode the given bytes to errs.Err,
// returns a standard error instead if unmarshal failed.
func Decode(body []byte) error {
	var e *Error
	err := json.Unmarshal(body, &e)
	if err != nil {
		return errors.New(string(body))
	}
	return e
}
