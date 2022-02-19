package podops

import "errors"

const (
	MsgStatus = "status: %d"

	// messages used in the validations
	MsgResourceInvalidName      = "invalid resource name '%s'"
	MsgResourceInvalidReference = "invalid reference type '%s'"
	MsgResourceIsInvalid        = "invalid resource '%s'"
	MsgResourceInvalidGUID      = "resource '%s': invalid guid '%s'" // FIXME swap order and update code accordingly -> "invalid guid '%s' for resource '%s'"
	MsgInvalidEmail             = "invalid email '%s'"
	MsgMissingCategory          = "missing categories"

	MsgResourceUnsupportedKind = "unsupported kind '%s'"
	MsgResourceImportError     = "error transfering '%s'"
	MsgResourceUploadError     = "error uploading '%s'"

	// CLI messages
	//MsgArgumentMissing       = "missing argument '%s'"
	//MsgTooManyArguments      = "too many arguments"
	//MsgArgumentCountMismatch = "argument mismatch: expected %d, got %d"

	MsgBuildSuccess    = "Sucessfully built podcast '%s'"
	MsgAssembleSuccess = "Sucessfully collected all resources"
	MsgGenerateSuccess = "Sucessfully generated markdown resources"
	MsgSyncSuccess     = "Sucessfully synced all resources"

	MsgConfigInit = "Created new config for client id '%s' with token '%s'"
	//MsgSecret     = "Refreshed the payload secret for podcast '%s': %s"
)

var (
	// ErrNotImplemented indicates that a function is not yet implemented
	//ErrNotImplemented = errors.New("not implemented")
	// ErrInternalError indicates everything else
	ErrInternalError = errors.New("internal error")
	// ErrApiError indicates an error in an API call
	ErrApiError = errors.New("api error")
	// ErrInvalidRoute indicates that the route and/or its parameters are not valid
	ErrInvalidRoute = errors.New("invalid route")
	// ErrMissingPayloadSecret indicates that no secret was provided
	//ErrMissingPayloadSecret = errors.New("missing payload secret")
	// ErrUnsupportedWebhookEvent indicates that the wrong type of webhook was received
	ErrUnsupportedWebhookEvent = errors.New("unsupported webhook")

	// ErrInvalidResourceName indicates that the resource name is invalid
	ErrInvalidResourceName = errors.New("invalid resource name")
	// ErrMissingResourceName indicates that a resource type is missing
	ErrMissingResourceName = errors.New("missing resource type")
	// ErrResourceNotFound indicates that the resource does not exist
	ErrResourceNotFound = errors.New("resource does not exist")
	// ErrResourceExists indicates that the resource does not exist
	//ErrResourceExists = errors.New("resource already exists")
	// ErrInvalidGUID indicates that the GUID is invalid
	ErrInvalidGUID = errors.New("invalid GUID")
	// ErrInvalidParameters indicates that parameters used in an API call are not valid
	ErrInvalidParameters = errors.New("invalid parameters")
	// ErrInvalidNumArguments indicates that the number of arguments in an API call is not valid
	ErrInvalidNumArguments = errors.New("invalid arguments")
	// ErrInvalidPassPhrase indicates that the pass phrase is too short
	ErrInvalidPassPhrase = errors.New("invalid pass phrase")

	// ErrBuildFailed indicates that there was an error while building the feed
	ErrBuildFailed = errors.New("build failed")
	// ErrBuildNoShow indicates that no show.yaml could be found
	ErrBuildNoShow = errors.New("missing show.yaml")
	// ErrBuildNoEpisodes indicates that no episodes could be found
	ErrBuildNoEpisodes = errors.New("missing episodes")

	// ErrAssembleNoResources indicates that no resources could be found
	ErrAssembleNoResources = errors.New("missing resource cache")
)
