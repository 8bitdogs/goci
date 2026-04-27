package payload

import (
	"encoding/json"
	"time"
)

type Time uint64

func (t *Time) UnmarshalJSON(b []byte) error {
	var i uint64
	err := json.Unmarshal(b, &i)
	if err != nil {
		// try to unmarshal as time.Time
		var tm time.Time
		err = json.Unmarshal(b, &tm)
		if err != nil {
			return err
		}
		*t = Time(tm.Unix())
		return nil
	}
	*t = Time(i)
	return nil
}
