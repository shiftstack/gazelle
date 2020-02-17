# go-sequence

Builds a slice of positive integers out of its string representation.
Sequence is expected to be a comma-separated list of intervals. An interval is either a number, or a pair of numbers separated by a dash.

Example:

   "1,3,6-8" -> []int{1, 3, 6, 7, 8}
