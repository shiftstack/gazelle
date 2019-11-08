package job

import (
	"encoding/json"
	"time"
)

// metadata contains the data parsed from "started.json" or "finished.json".
type metadata struct {
	time   time.Time
	result string
}

// UnmarshalJSON implements json.Unmarshal for metadata. The purpose of this
// method is to parse a time.Time out of an epoch timestamp.
func (m *metadata) UnmarshalJSON(src []byte) error {
	var data struct {
		Timestamp int64
		Result    string
	}
	err := json.Unmarshal(src, &data)
	if err != nil {
		return err
	}

	t := time.Unix(data.Timestamp, 0)

	*m = metadata{
		time:   t.In(time.UTC),
		result: data.Result,
	}

	return nil
}
