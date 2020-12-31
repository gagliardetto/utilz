package utilz

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gagliardetto/hashsearch"
	"github.com/miekg/dns"
)

type Replacer struct {
	s string
	n int
}

func NewReplacer(s string, n int) *Replacer {
	return &Replacer{
		s: s,
		n: n,
	}
}

//
func (re *Replacer) Replace(old, new string) *Replacer {
	re.s = strings.Replace(re.s, old, new, re.n)
	return re
}

func (re *Replacer) String() string {
	return re.s
}

var illegalFileNameCharacters = regexp.MustCompile(`[^[a-zA-Z0-9]-_]`)

func SanitizeFileNamePart(part string) string {
	part = strings.Replace(part, "/", "-", -1)
	part = illegalFileNameCharacters.ReplaceAllString(part, "")
	return part
}

// SliceContains returns true if the provided slice of strings contains the element
func SliceContains(slice []string, element string) bool {
	for _, elem := range slice {
		if strings.EqualFold(element, elem) {
			return true
		}
	}
	return false
}

// IntSliceContains returns true if the provided slice of ints contains the element
func IntSliceContains(slice []int, element int) bool {
	for _, elem := range slice {
		if element == elem {
			return true
		}
	}
	return false
}

// ConcatBytesFromStrings concatenates multiple strings into one array of bytes
func ConcatBytesFromStrings(strs ...string) ([]byte, error) {
	var concatBytesBuffer bytes.Buffer
	for _, stringValue := range strs {
		_, err := concatBytesBuffer.WriteString(stringValue)
		if err != nil {
			return []byte{}, err
		}
	}
	return concatBytesBuffer.Bytes(), nil
}

// NewUniqueElements removes elements that have duplicates in the original or new elements.
func NewUniqueElements(orig []string, add ...string) []string {
	var n []string

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

// UniqueAppend behaves like the Go append, but does not add duplicate elements.
func UniqueAppend(orig []string, add ...string) []string {
	return append(orig, NewUniqueElements(orig, add...)...)
}

// Deduplicate returns a deduplicated copy of a.
func Deduplicate(a []string) []string {
	var res []string
	res = UniqueAppend(res, a...)
	return res
}

// DeduplicateSlice returns a deduplicated copy of provided slice.
func DeduplicateSlice(slice interface{}, fn func(i int) string) interface{} {
	// TODO: improve and test this func
	storeIndex := hashsearch.New()

	var result reflect.Value

	switch reflect.TypeOf(slice).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(slice)

		myType := reflect.TypeOf(slice).Elem()
		result = reflect.MakeSlice(reflect.SliceOf(myType), 0, 0)

		for i := 0; i < s.Len(); i++ {
			id := fn(i)
			if !storeIndex.Has(id) {
				result = reflect.Append(result, s.Index(i))
				storeIndex.OrderedAppend(id)
			}
		}
	default:
		panic(Sf("%s is not a slice, but a %s", reflect.TypeOf(slice).Name(), reflect.TypeOf(slice).Kind()))
	}
	return result.Interface()
}

// DeduplicateSlice2 deduplicates a slice by modifying it.
func DeduplicateSlice2(slice interface{}, fn func(i int) string) {
	// TODO: improve and test this func

	storeIndex := hashsearch.New()

	if reflect.TypeOf(slice).Kind() != reflect.Ptr {
		panic(Sf("Expected a pointer to a slice, but got a %s", reflect.TypeOf(slice).Kind()))
	}

	ptrElemKind := reflect.TypeOf(slice).Elem().Kind()
	switch ptrElemKind {
	case reflect.Slice:
		rv := reflect.Indirect(reflect.ValueOf(slice))

		indexesToRemove := make([]int, 0)

		// Gather indexes of duplicate elements that will be removed:
		for i := 0; i < rv.Len(); i++ {
			id := fn(i)
			if !storeIndex.Has(id) {
				storeIndex.OrderedAppend(id)
			} else {
				indexesToRemove = append(indexesToRemove, i)
			}
		}

		// Create an update-able map of indexes:
		im := make(map[int]int)
		for _, i := range indexesToRemove {
			im[i] = i
		}
		// Remove duplicates:
		for _, i := range indexesToRemove {
			ri := im[i]
			deleteElementAtIndexFromSlice(rv, ri, im)
		}
	default:
		panic(Sf("Expected a pointer to a slice, but got a pointer to a %s", ptrElemKind))
	}
}

//
func deleteElementAtIndexFromSlice(rv reflect.Value, index int, im map[int]int) bool {
	if index > rv.Len()-1 {
		return false
	}

	zero := reflect.Zero(reflect.TypeOf(rv.Interface()).Elem())

	lastIndex := rv.Len() - 1
	if _, ok := im[lastIndex]; ok {
		im[lastIndex] = index // Update index of last element.
	}
	// Remove the element at index i from rv.
	rv.Index(index).Set(rv.Index(lastIndex)) // Copy last element to index i.
	rv.Index(lastIndex).Set(zero)            // Erase last element (write zero value).
	rv.Set(rv.Slice(0, lastIndex))           // Truncate slice.
	return true
}

func HasDuplicate(key string, slice interface{}, fn func(i int) string) bool {
	// TODO: improve and test this func

	switch reflect.TypeOf(slice).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(slice)

		for i := 0; i < s.Len(); i++ {
			out := fn(i)
			if out == key {
				return true
			}
		}
	default:
		panic(Sf("%s is not a slice, but a %s", reflect.TypeOf(slice).Name(), reflect.TypeOf(slice).Kind()))
	}
	return false
}

type ElasticStringIterator struct {
	arr []string
}

func NewElasticStringIterator(arr []string) *ElasticStringIterator {
	return &ElasticStringIterator{
		arr: arr,
	}
}

func (e *ElasticStringIterator) Iterate(callback func(s string) ([]string, bool)) {
	tot := len(e.arr)
	for i := 0; i < tot; i++ {
		x := e.arr[i]
		toAdd, goOn := callback(x)
		if !goOn {
			break
		}
		if toAdd != nil {
			for _, addThis := range toAdd {
				e.arr = append(e.arr, addThis)
				tot++
			}
		}
	}
}

func GetAddedRemoved(previous []string, next []string) (added []string, removed []string) {
	for _, prev := range previous {
		if !SliceContains(next, prev) {
			removed = append(removed, prev)
		}
	}

	for _, nx := range next {
		if !SliceContains(previous, nx) {
			added = append(added, nx)
		}
	}

	return
}

var ConstantPredeterminedLength = 60

func CustomConstantLength(ln int, s string) string {
	if len(s) < ln {
		s += ReturnNSpaces(ln - len(s))
		return s
	}
	return s
}

func ConstantLength(s string) string {
	if len(s) < ConstantPredeterminedLength {
		s += ReturnNSpaces(ConstantPredeterminedLength - len(s))
		return s
	}
	return s
}
func RepeatString(n int, char string) string {
	var res string
	for i := 0; i < n; i++ {
		res += char
	}
	return res
}
func ReturnNSpaces(n int) string {
	return RepeatString(n, " ")
}
func IndentWithSpaces(n int, s string) string {
	return ReturnNSpaces(n) + s
}
func NewSpacesIndenter(spaceWidth int) func(n int, s string) string {
	return func(n int, s string) string {
		return ReturnNSpaces(n*spaceWidth) + s
	}
}

func IndentWithTabs(n int, s string) string {
	return RepeatString(n, "	") + s
}
func IsNotAnyOf(s string, candidates ...string) bool {
	return !IsAnyOf(s, candidates...)
}
func IsAnyOf(s string, candidates ...string) bool {
	for _, v := range candidates {
		if s == v {
			return true
		}
	}
	return false
}

// Reverses a string
func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// ReverseStringSlice reverses a string slice
func ReverseStringSlice(ss []string) {
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}
}

// ReverseDNSLabels reverses the labels of a DNS, and returns the DNS name.
func ReverseDNSLabels(hostname string) string {
	labels := dns.SplitDomainName(hostname)
	ReverseStringSlice(labels)
	reversed := strings.Join(labels, ".")
	return reversed
}

// HasAnySuffixDottedOrNot returns true if the string s has any of the
// suffixes, whether they are in the .<suffix> or <suffix> variant.
func HasAnySuffixDottedOrNot(s string, suffixes ...string) bool {
	newSuffixes := make([]string, 0)

	for _, suffixCandidate := range suffixes {
		suffix := strings.TrimPrefix(suffixCandidate, ".")
		newSuffixes = append(newSuffixes,
			suffix,
			"."+suffix,
		)
	}

	return HasAnySuffix(s, newSuffixes...)
}

// HasAnySuffix returns true if the string s has any of the prefixes
// provided.
func HasAnySuffix(s string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}

// TrimDNSLabelFromBeginning removes one label from the DNS name,
// and returns the new DNS name and the removed label.
func TrimDNSLabelFromBeginning(name string) (string, string) {
	labels := dns.SplitDomainName(name)

	if len(labels) == 0 {
		// TODO: what to return???
		return "", name
	}
	labels = labels[1:]

	return strings.Join(labels, "."), labels[0]
}

func InitDomainExtensions(path string) {
	domainExtensions = func() []string {
		extensions := make([]string, 0)

		err := ReadFileLinesAsString(
			path,
			func(ext string) bool {
				extensions = append(extensions, ext)
				return true
			})

		if err != nil {
			panic(err)
		}

		// sort by length:
		sort.Strings(extensions)
		// invert to have the longer extensions first:
		sort.Sort(sort.Reverse(sort.StringSlice(extensions)))
		return extensions
	}()
}

var domainExtensions []string

// GetDomainExtension returns the domain extension (among the known ones)
// of the provided host names.
func GetDomainExtension(name string) string {
	for _, ext := range domainExtensions {
		if strings.HasSuffix(name, ext) {
			return ext
		}
	}
	return ""
}

// GetBaseDomain returns the first label after the domain extension,
// ad the domain extension.
//
// e.g. example.com -> example + .com
// e.g. subdomain.example.com -> example + .com
//
// e.g. example.co.uk -> example + .co.uk
// e.g. subdomain.example.co.uk -> example + .co.uk
func GetBaseDomain(name string) (string, string) {
	ext := GetDomainExtension(name)
	if ext == "" {
		panic(Sf("no extension found for %q", name))
	}
	hostPart := strings.TrimSuffix(name, ext)
	trimmed := hostPart

	for {
		remainingLabels := dns.CountLabel(hostPart)
		if remainingLabels > 1 {
			hostPart, trimmed = TrimDNSLabelFromBeginning(hostPart)
		} else {
			break
		}
	}
	return trimmed, ext
}

// TrimFrom removes the string `after` from the string `s`
// and anything that comes after it.
func TrimFrom(s string, after string) string {
	parts := strings.Split(s, after)
	if len(parts) > 0 {
		return parts[0]
	}
	return s
}

// FilterModify returns a new array containing the values of `ss` after they
// have been processed by the func `filter`; no modification to `ss` is done.
func FilterModify(ss []string, filter func(s string) string) []string {
	filtered := make([]string, 0)
	for _, s := range ss {
		filtered = append(filtered, filter(s))
	}
	return filtered
}

// FilterExclude returns a new array of items that do not match the filter func;
// for any item in `ss`, if the `filter` func returns true,
// then the item will be excluded from the result slice.
func FilterExclude(ss []string, filter func(s string) bool) []string {
	filtered := make([]string, 0)
	for _, s := range ss {
		exclude := filter(s)
		if !exclude {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func CloneSlice(sl []string) []string {
	clone := make([]string, len(sl))
	copy(clone, sl)
	return clone
}

func SplitStringSlice(parts int, slice []string) [][]string {
	var divided [][]string

	chunkSize := (len(slice) + parts - 1) / parts

	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		if end > len(slice) {
			end = len(slice)
		}

		divided = append(divided, slice[i:end])
	}

	return divided
}
func AnyIsEmptyString(slice ...string) bool {
	for _, v := range slice {
		if v == "" {
			return true
		}
	}
	return false
}

func IsLastUnicode(s string) bool {
	return s == "10FFFF"
}

func MustParseUnicode(code string) string {
	parsed, err := ParseUnicode(code)
	if err != nil {
		panic(fmt.Errorf("error for %s: %s", code, err))
	}
	return string(parsed)
}
func MustParseUnicodeAsRune(code string) rune {
	parsed, err := ParseUnicode(code)
	if err != nil {
		panic(fmt.Errorf("error for %s: %s", code, err))
	}
	return parsed
}

func ParseUnicode(code string) (rune, error) {
	// Trim eventual `\u` prefix:
	code = strings.TrimPrefix(code, `\u`)

	unquoted, _, _, err := strconv.UnquoteChar(`\u`+code, '\'')
	if err != nil {
		return ' ', err
	}
	return unquoted, nil
}
func SplitStringByRune(s string) []rune {
	var res []rune
	IterateStringAsRunes(s, func(r rune) bool {
		res = append(res, r)
		return true
	})
	return res
}
func SplitStringByRuneAsStringSlice(s string) []string {
	var res []string
	IterateStringAsRunes(s, func(r rune) bool {
		res = append(res, string(r))
		return true
	})
	return res
}
func IterateStringAsRunes(s string, callback func(r rune) bool) {
	for _, rn := range s {
		//fmt.Printf("%d: %c\n", i, rn)
		doContinue := callback(rn)
		if !doContinue {
			return
		}
	}
}
func ToLower(s string) string {
	return strings.ToLower(s)
}
func ToTitle(s string) string {
	return strings.ToTitle(s)
}

// Map maps the provided slice.
func Map(slice interface{}, fn func(i int) string) []string {
	var res []string
	switch reflect.TypeOf(slice).Kind() {
	case reflect.Slice:
		rv := reflect.ValueOf(slice)

		for i := 0; i < rv.Len(); i++ {
			res = append(res, fn(i))
		}
	default:
		panic(Sf("Expected a slice, but got a %s", reflect.TypeOf(slice).Kind()))
	}
	return res
}
