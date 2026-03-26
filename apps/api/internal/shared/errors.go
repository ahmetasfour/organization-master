package shared

import "errors"

// Sentinel errors for common error conditions
var (
	// ErrApplicationTerminated indicates an application has been terminated and cannot be modified
	ErrApplicationTerminated = errors.New("application has been terminated and cannot be modified")

	// ErrTokenNotFound indicates the token was not found in the database
	ErrTokenNotFound = errors.New("token not found")

	// ErrTokenExpired indicates a token has expired
	ErrTokenExpired = errors.New("token has expired")

	// ErrTokenUsed indicates a token has already been used
	ErrTokenUsed = errors.New("token has already been used")

	// ErrInvalidCredentials indicates invalid login credentials
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrForbidden indicates the user lacks permission for the requested action
	ErrForbidden = errors.New("you do not have permission to perform this action")

	// ErrNotFound indicates the requested resource was not found
	ErrNotFound = errors.New("resource not found")

	// ErrDuplicateVote indicates a user has already voted in a stage
	ErrDuplicateVote = errors.New("you have already voted in this stage")

	// ErrInvalidTransition indicates an invalid state transition
	ErrInvalidTransition = errors.New("invalid state transition")

	// ErrReferenceQuotaExceeded indicates too many references have been requested
	ErrReferenceQuotaExceeded = errors.New("reference quota exceeded")

	// ErrUnauthorized indicates the request lacks valid authentication
	ErrUnauthorized = errors.New("unauthorized")
)
