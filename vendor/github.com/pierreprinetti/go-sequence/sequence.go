package sequence

import (
	"strconv"
	"strings"
)

// Int builds a slice of positive integers out of its string representation.
// Sequence is expected to be a comma-separated list of intervals. An interval is either a number, or a pair of numbers separated by a dash.
//
// Example:
//
//    "1,3,6-8" -> []int{1, 3, 6, 7, 8}
func Int(sequence string) ([]int, error) {
	var seq []int
	if sequence == "" {
		return seq, nil
	}

	for _, interval := range strings.Split(sequence, ",") {
		intervalBoundaries := strings.SplitN(interval, "-", 2)

		if len(intervalBoundaries) < 2 {
			n, err := strconv.Atoi(strings.TrimSpace(intervalBoundaries[0]))
			if err != nil {
				return nil, err
			}
			seq = append(seq, n)
			continue
		}

		start, err := strconv.Atoi(strings.TrimSpace(intervalBoundaries[0]))
		if err != nil {
			return nil, err
		}

		end, err := strconv.Atoi(strings.TrimSpace(intervalBoundaries[1]))
		if err != nil {
			return nil, err
		}

		if end >= start {
			for n := start; n <= end; n++ {
				seq = append(seq, n)
			}
		} else {
			for n := start; n >= end; n-- {
				seq = append(seq, n)
			}
		}
	}
	return seq, nil
}
