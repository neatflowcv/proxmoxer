package common

import "errors"

// Domain layer errors
var (
	ErrClusterNotFound       = errors.New("cluster not found")
	ErrClusterAlreadyExists  = errors.New("cluster already exists")
	ErrInvalidClusterID      = errors.New("invalid cluster id")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrAuthenticationFailed  = errors.New("authentication failed")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrInternalError         = errors.New("internal error")
	ErrProxmoxConnectionFailed = errors.New("failed to connect to proxmox")
)
