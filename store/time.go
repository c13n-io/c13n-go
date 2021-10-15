package store

import "time"

func generateTimestamp() time.Time {
	return time.Now()
}

var getCurrentTime = generateTimestamp
