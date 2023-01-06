package job

import (
	"encoding/json"
	"time"
)

type Finished struct {
	Time   time.Time
	Passed bool
	// Metadata json.RawMessage
	Result   string
	Revision string
}

func (f *Finished) UnmarshalJSON(src []byte) error {
	var source struct {
		Timestamp int64  `json:"timestamp"`
		Passed    bool   `json:"passed"`
		Result    string `json:"result"`
		Revision  string `json:"revision"`
	}
	if err := json.Unmarshal(src, &source); err != nil {
		return err
	}

	t := time.Unix(source.Timestamp, 0)

	*f = Finished{
		Time:     t,
		Passed:   source.Passed,
		Result:   source.Result,
		Revision: source.Revision,
	}
	return nil
}
