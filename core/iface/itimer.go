package iface

import "time"

// ITimer interface for timer operations
type ITimer interface {
	// Stop stops the timer
	// Returns false if the timer has already been triggered or stopped
	Stop() bool

	// Reset resets the timer
	// interval: new interval duration (if 0, uses the existing interval)
	// Returns whether the reset was successful
	Reset(interval time.Duration) bool

	// IsActive checks if the timer is active
	IsActive() bool

	// Interval gets the current interval duration
	Interval() time.Duration

	// NextTrigger gets the next trigger time
	NextTrigger() time.Time

	// Execute executes the timer callback
	Execute() error
}
