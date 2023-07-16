package core

import "errors"

var (
	// DB
	ErrDbUrlIsRequired         = errors.New("DB_URI environment variable is required")
	ErrDbInvalidDatabaseEngine = errors.New("Invalid database engine")

	// View
	ErrTemplateNotFound = errors.New("Template not found")
)
