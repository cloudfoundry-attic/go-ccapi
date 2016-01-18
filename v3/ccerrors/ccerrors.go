package ccerrors

import "errors"

// ErrInvalidToken is returned when any client API call
// fails due to the provided token being invalid/expired.
var ErrInvalidToken = errors.New("ErrInvalidToken")
