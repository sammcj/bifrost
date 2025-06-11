package bifrost

import schemas "github.com/maximhq/bifrost/core/schemas"

func Ptr[T any](v T) *T {
	return &v
}

// newBifrostError wraps a standard error into a BifrostError with IsBifrostError set to false.
// This helper function reduces code duplication when handling non-Bifrost errors.
func newBifrostError(err error) *schemas.BifrostError {
	return &schemas.BifrostError{
		IsBifrostError: false,
		Error: schemas.ErrorField{
			Message: err.Error(),
			Error:   err,
		},
	}
}

// newBifrostErrorFromMsg creates a BifrostError with a custom message.
// This helper function is used for static error messages.
func newBifrostErrorFromMsg(message string) *schemas.BifrostError {
	return &schemas.BifrostError{
		IsBifrostError: false,
		Error: schemas.ErrorField{
			Message: message,
		},
	}
}
