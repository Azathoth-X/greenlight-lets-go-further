package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Runtime int32

var ErrorInvalidRuntimeFormat = errors.New("Invalid Runtime Format")

func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)

	quotedJson := strconv.Quote(jsonValue)

	return []byte(quotedJson), nil

}

func (r *Runtime) UnMarshallJSON(jsonValue []byte) error {

	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))

	if err != nil {
		return ErrorInvalidRuntimeFormat
	}

	splitStrings := strings.Split(unquotedJSONValue, "")

	if len(splitStrings) < 2 || splitStrings[1] != "mins" {
		return ErrorInvalidRuntimeFormat
	}

	i, err := strconv.ParseInt(splitStrings[0], 10, 32)

	if err != nil {
		return ErrorInvalidRuntimeFormat
	}

	*r = Runtime(i)

	return nil
}
