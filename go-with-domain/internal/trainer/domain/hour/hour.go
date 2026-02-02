package hour

import "time"

type Hour struct {
	hour time.Time

	availability Availability
}