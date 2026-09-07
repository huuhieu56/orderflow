package notification

import "errors"

var ErrNotificationNotFound = errors.New("notification not found")
var ErrInvalidEvent = errors.New("invalid event")
