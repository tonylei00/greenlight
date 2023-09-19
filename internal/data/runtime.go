package data

import (
	"fmt"
	"strconv"
)

type Runtime int32

func (r Runtime) MarshalJSON() ([]byte, error) {
	formatted := fmt.Sprintf("%d mins", r)

	// Wraps the string in double quotes in order to be a valid *JSON string*
	json := strconv.Quote(formatted)

	return []byte(json), nil
}
