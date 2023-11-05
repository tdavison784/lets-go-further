package data

import (
	"fmt"
	"strconv"
)

// Runtime Declare a custom runtime type, which has the underlying type int32
type Runtime int32

// MarshalJSON Implement a MarshalJSON() method on the Runtime type so that it satisfies the json.Marshal interface
// this should return the json-encoded value for the movie runtime (in our case it will return a string in the format
// "<runtime> mins").
func (r Runtime) MarshalJSON() ([]byte, error) {
	// Generate a string containing the move runtime in the required format
	jsonValue := fmt.Sprintf("%d mins", r)

	// Use strconv.Quote() function on the string to wrap it in double quotes.
	// It needs to be surrounded by double quotes in order to be a valid *JSON string*
	quotedJSONValue := strconv.Quote(jsonValue)

	return []byte(quotedJSONValue), nil
}
