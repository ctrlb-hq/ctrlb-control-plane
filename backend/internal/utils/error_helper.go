package utils

import "errors"

var ErrUserAlreadyExists = errors.New("user already exists")

var ErrAgentDoesNotExists = errors.New("agent doesn't exist")

var ErrPipelineDoesNotExists = errors.New("pipeline doesn't exist")

var ErrInvalidConfig = errors.New("agent returned 500 - invalid config")
