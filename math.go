package utilz

import "strconv"

// Percent calculate what is [percent]% of [number]
// For Example 25% of 200 is 50
// It returns result as float64
func Percent(pcent int64, all int64) float64 {
	percent := ((float64(all) * float64(pcent)) / float64(100))
	return percent
}

// PercentOf calculate [number1] is what percent of [number2]
// For example 300 is 12.5% of 2400
// It returns result as float64
func PercentOf(current int64, all int64) float64 {
	if current == 0 || all == 0 {
		return 0.0
	}
	percent := (float64(current) * float64(100)) / float64(all)
	return percent
}
func GetPercent(done int64, all int64) float64 {
	if all == 0 || done == 0 {
		return 0.0
	}
	percent := float64(100) / (float64(all) / float64(done))
	return percent
}
func GetFormattedPercent(done int64, all int64) string {
	percentDone := Sf("%s%%", strconv.FormatFloat(GetPercent(done, all), 'f', 2, 64))
	return percentDone
}

// Change calculate what is the percentage increase/decrease from [number1] to [number2]
// For example 60 is 200% increase from 20
// It returns result as float64
func Change(before int64, after int64) float64 {
	diff := float64(after) - float64(before)
	realDiff := diff / float64(before)
	percentDiff := 100 * realDiff

	return percentDiff
}

// generate number range including min and max:
func GenerateIntRangeInclusive(min, max int) []int {
	ints := []int{}
	for i := min; i <= max; i++ {
		ints = append(ints, i)
	}
	return ints
}

// NewUniqueInts removes elements that have duplicates in the original or new elements.
func NewUniqueInts(orig []int, add ...int) []int {
	var n []int

	for _, av := range add {
		found := false
		s := av

		// Check the original slice for duplicates
		for _, ov := range orig {
			if s == ov {
				found = true
				break
			}
		}
		// Check that we didn't already add it in
		if !found {
			for _, nv := range n {
				if s == nv {
					found = true
					break
				}
			}
		}
		// If no duplicates were found, add the entry in
		if !found {
			n = append(n, s)
		}
	}
	return n
}

// UniqueAppendInts behaves like the Go append, but does not add duplicate elements.
func UniqueAppendInts(orig []int, add ...int) []int {
	return append(orig, NewUniqueInts(orig, add...)...)
}
func DeduplicateInts(a []int) []int {
	var res []int
	res = UniqueAppendInts(res, a...)
	return res
}
