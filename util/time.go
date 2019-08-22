package util

import "time"

//to replace time when testing
var TimeNow = func() time.Time {
	return time.Now()
}
