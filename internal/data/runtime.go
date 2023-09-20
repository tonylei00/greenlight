package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Runtime int32

var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

// Encodes int32 to "<int32> mins"
func (r Runtime) MarshalJSON() ([]byte, error) {
	formatted := fmt.Sprintf("%d mins", r)

	// Wraps the string in double quotes in order to be a valid *JSON string*
	json := strconv.Quote(formatted)

	return []byte(json), nil
}

// Decodes "<int32> mins" to int32
func (r *Runtime) UnmarshalJSON(input []byte) error {
	// 1. Unquote the double quotes from the JSON string
	unquotedJSONStr, err := strconv.Unquote(string(input))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// 2. Split the string and do a sanity check
	parts := strings.Split(unquotedJSONStr, " ")
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// 3. Attempt to convert the runtime value to an int32
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// 4. Deference the pointer receiver and set its underlying value to type of Runtime
	*r = Runtime(i)

	return nil
}
