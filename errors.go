package core

import "errors"

var (
	// DB
	ErrDbUrlIsRequired         = errors.New("DB_URI environment variable is required")
	ErrDbInvalidDatabaseEngine = errors.New("invalid database engine")

	// View
	ErrTemplateNotFound = errors.New("template not found")
)
