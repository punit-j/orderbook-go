package errors

import "errors"

var (
	// DB Errors
	ErrFailedToConnectDB = errors.New("failed to connect to database")
	ErrFindingDriver = errors.New("no migrate driver instance found")
	ErrUnableToRetrieveOrder = errors.New("unable to retrieve orders from db")

	// Vaildation Errors
	ErrMissingTrader = errors.New("missing order trader")
	ErrNilOrder = errors.New("nil order, illegal entry")
	ErrMissingAssets = errors.New("missing order assets")
)
