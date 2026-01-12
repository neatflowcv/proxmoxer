package common

import "errors"

// Domain layer errors.
var (
	ErrClusterNotFound         = errors.New("cluster not found")
	ErrClusterAlreadyExists    = errors.New("cluster already exists")
	ErrInvalidClusterID        = errors.New("invalid cluster id")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrAuthenticationFailed    = errors.New("authentication failed")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrInternalError           = errors.New("internal error")
	ErrProxmoxConnectionFailed = errors.New("failed to connect to proxmox")
	ErrRequestNil              = errors.New("request cannot be nil")
	ErrClusterNameRequired     = errors.New("cluster name is required")
	ErrClusterNameTooLong      = errors.New("cluster name must be at most 255 characters")
	ErrAPIEndpointRequired     = errors.New("api endpoint is required")
	ErrUsernameRequired        = errors.New("username is required")
	ErrPasswordRequired        = errors.New("password is required")
	ErrClusterIDEmpty          = errors.New("cluster id cannot be empty")
	ErrClusterNameEmpty        = errors.New("cluster name cannot be empty")
	ErrAPIEndpointEmpty        = errors.New("api endpoint cannot be empty")
	ErrUsernameEmpty           = errors.New("username cannot be empty")
	ErrPasswordEmpty           = errors.New("password cannot be empty")
	ErrClusterNil              = errors.New("cluster cannot be nil")
	ErrNoAuthenticationTicket  = errors.New("no authentication ticket received")
	ErrDiskQueryFailed         = errors.New("failed to query disk information")
)
