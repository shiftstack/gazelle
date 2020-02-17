# go-junit

Go types for unmarshaling JUnit XML files.

## Example

```
import (
	"fmt"

	"github.com/pierreprinetti/go-junit"
)

func main() {
	f, err := os.Open("junit_suite1.xml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var testSuite junit.TestSuite

	if err := xml.NewDecoder(r).Decode(&testSuite); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	for _, tc := range testSuite.TestCases {
		if tc.Failure != nil {
			fmt.Printf("FAILED: %s\n", tc.Name)
			continue
		}

		if tc.SkipMessage != nil {
			fmt.Printf("SKIPPED: %s\n", tc.SkipMessage.Message)
			continue
		}
		fmt.Printf("PASS: %s\n", tc.Name)
	}
}
```

## Credit

This work is based on [jstemmer/go-junit-report](https://github.com/jstemmer/go-junit-report), which is Copyright (c) 2012 Joel Stemmer.
