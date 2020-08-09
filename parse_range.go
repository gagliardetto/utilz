package utilz

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseIntervals parses a string representing a list of unsigned integers or integer ranges (using dashes).
func ParseIntervals(s string) ([]int, error) {
	// trim space:
	s = strings.TrimSpace(s)
	// remove all spaces:
	s = strings.ReplaceAll(s, " ", "")
	// split by comma:
	commaSeparated := strings.Split(s, ",")

	var results []int
	for _, rr := range commaSeparated {
		dashCount := strings.Count(rr, "-")
		switch dashCount {
		case 0:
			{
				parsed, err := strconv.Atoi(rr)
				if err != nil {
					return nil, fmt.Errorf("error parsing %q: %s", rr, err)
				}
				results = append(results, parsed)
			}
		case 1:
			{
				start, end, err := parseRange(rr)
				if err != nil {
					return nil, err
				}
				rangeValues := GenerateIntRangeInclusive(start, end)
				results = append(results, rangeValues...)
			}
		default:
			{
				return nil, fmt.Errorf("error: %q contains too many dashes", rr)
			}
		}
	}

	return results, nil
}

func parseRange(s string) (int, int, error) {
	rangeVals := strings.Split(s, "-")
	if len(rangeVals) != 2 {
		return 0, 0, fmt.Errorf("cannot parse range: %q", s)
	}

	rangeStartString := rangeVals[0]
	start, err := strconv.Atoi(rangeStartString)
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing range start %q: %s", rangeStartString, err)
	}

	rangeEndString := rangeVals[1]
	end, err := strconv.Atoi(rangeEndString)
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing range end %q: %s", rangeEndString, err)
	}

	return start, end, nil
}
func example() {
	res, err := ParseIntervals("99,5,1-3,98-90")
	if err != nil {
		panic(err)
	}

	fmt.Println(res)
	fmt.Println(DeduplicateInts(res))
}
