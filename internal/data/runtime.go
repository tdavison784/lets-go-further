package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrInvalidRuntimeFormat define an error that our UnmarshalJSON method can return if we're unable to parse
// or convert the JSON string successfully
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

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

// UnmarshalJSON method on the Runtime type so that it satisfies the json.Unmarshaler interface.
// IMPORTANT: Because UnmarshalJSON() needs to modify the receiver (our Runtime type), we must
// use a pointer receiver for this to work correctly. Otherwise, we wil only be modifying a copy
// which is then discarded when this method returns
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	// We expect the incoming JSON value will be a sring in the format
	// "<runtime> mins", and the first thing we need to do is remove the surrounding
	// double quotes from this string. If we can't unquote it, then we return the
	// ErrInvalidRuntimeFormat error
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Split the string to isolate the part containing the number
	parts := strings.Split(unquotedJSONValue, " ")

	// Sanity check the parts of the string to make sure it was in the expected format.
	// if it isn't, we return the ErrInvalidRuntimeFormat error again
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// Otherwise, parse the string containing the number into an int32. Again, if this fails
	// return ErrInvalidRuntimeFormat error
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}
	*r = Runtime(i)
	return nil
}
