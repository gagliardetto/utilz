package utilz

import (
	"fmt"
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/hashsearch"
)

// DeduplicateSlice returns a deduplicated copy of the provided slice.
func DeduplicateSlice(slice interface{}, fn func(i int) string) interface{} {
	// TODO: improve and test this func
	storeIndex := hashsearch.New()

	var result reflect.Value

	switch reflect.TypeOf(slice).Kind() {
	case reflect.Slice, reflect.Array:
		rv := reflect.ValueOf(slice)

		myType := reflect.TypeOf(slice).Elem()
		result = reflect.MakeSlice(reflect.SliceOf(myType), 0, 0)

		for i := 0; i < rv.Len(); i++ {
			id := fn(i)
			if !storeIndex.Has(id) {
				result = reflect.Append(result, rv.Index(i))
				storeIndex.OrderedAppend(id)
			}
		}
	default:
		panic(Sf("Expected a slice/array, but got a %s", reflect.TypeOf(slice).Kind()))
	}
	return result.Interface()
}

// DeduplicateSlice2 deduplicates a slice/array by modifying it.
// Must pass a pointer to a slice/array as first argument.
func DeduplicateSlice2(slicePtr interface{}, fn func(i int) string) {
	// TODO: improve and test this func

	storeIndex := hashsearch.New()

	if reflect.TypeOf(slicePtr).Kind() != reflect.Ptr {
		panic(Sf("Expected a pointer to a slice/array, but got a %s", reflect.TypeOf(slicePtr).Kind()))
	}

	ptrElemKind := reflect.TypeOf(slicePtr).Elem().Kind()
	switch ptrElemKind {
	case reflect.Slice, reflect.Array:
		rv := reflect.Indirect(reflect.ValueOf(slicePtr))

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
		panic(Sf("Expected a pointer to a slice/array, but got a pointer to a %s", ptrElemKind))
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
	case reflect.Slice, reflect.Array:
		s := reflect.ValueOf(slice)

		for i := 0; i < s.Len(); i++ {
			out := fn(i)
			if out == key {
				return true
			}
		}
	default:
		panic(Sf("Expected a slice or array, but got a %s", reflect.TypeOf(slice).Kind()))
	}
	return false
}

// MapSlice maps the provided slice.
func MapSlice(slice interface{}, fn func(i int) string) []string {
	var res []string
	switch reflect.TypeOf(slice).Kind() {
	case reflect.Slice, reflect.Array:
		rv := reflect.ValueOf(slice)

		for i := 0; i < rv.Len(); i++ {
			res = append(res, fn(i))
		}
	default:
		panic(Sf("Expected a slice or array, but got a %s", reflect.TypeOf(slice).Kind()))
	}
	return res
}

// FilterSlice filters the provided slice.
func FilterSlice(slice interface{}, fn func(i int) bool) interface{} {
	// TODO: improve and test this func
	var result reflect.Value

	switch reflect.TypeOf(slice).Kind() {
	case reflect.Slice, reflect.Array:
		rv := reflect.ValueOf(slice)

		myType := reflect.TypeOf(slice).Elem()
		result = reflect.MakeSlice(reflect.SliceOf(myType), 0, 0)

		for i := 0; i < rv.Len(); i++ {
			include := fn(i)
			if include {
				result = reflect.Append(result, rv.Index(i))
			}
		}
	default:
		panic(Sf("Expected a slice or array, but got a %s", reflect.TypeOf(slice).Kind()))
	}
	return result.Interface()
}

// Map maps a slice/array/map. The mapper must be a function
// with one parameter (an int for slices/arrays, or the map key type for maps),
// and one return argument.
// Map returns an array of the mapper return argument type.
func Map(elem interface{}, mapper interface{}) interface{} {
	res, err := mapWithFilter(elem, mapper, passThroughFilter)
	if err != nil {
		panic(err)
	}
	return res
}

// passThroughFilter allows all values.
func passThroughFilter(reflect.Value) bool {
	return true
}

// nilFilter excludes nil values.
func nilFilter(rv reflect.Value) bool {
	return rv.Interface() != nil
}

var emptyInt int
var integerType = reflect.ValueOf(emptyInt).Type()

func mapWithFilter(cont interface{}, mapper interface{}, filter func(reflect.Value) bool) (interface{}, error) {
	// mapper must be a func:
	if mapperKind := reflect.TypeOf(mapper).Kind(); mapperKind != reflect.Func {
		return nil, fmt.Errorf("mapper is not a func, but a %s", mapperKind)
	}
	ff := reflect.ValueOf(mapper)
	// mapper must have one parameter:
	if ff.Type().NumIn() != 1 {
		return nil, fmt.Errorf("wrong number of parameters for mapper func: want 1, got %v", ff.Type().NumIn())
	}
	// mapper must have one return value:
	if ff.Type().NumOut() != 1 {
		return nil, fmt.Errorf("wrong number of return arguments for mapper func: want 1, got %v", ff.Type().NumOut())
	}

	containerType := reflect.TypeOf(cont)
	inType := ff.Type().In(0)
	outType := ff.Type().Out(0)

	rv := reflect.ValueOf(cont)
	result := reflect.MakeSlice(reflect.SliceOf(outType), 0, 0)

	switch containerType.Kind() {
	case reflect.Slice, reflect.Array:
		if !reflect.DeepEqual(inType, integerType) {
			return nil, fmt.Errorf("mapper func arg must be an int, but got %s", inType.Kind())
		}
		// Iterate over the slice/array, and call the mapper:
		for index := 0; index < rv.Len(); index++ {
			returned := ff.Call([]reflect.Value{reflect.ValueOf(index)})[0]
			if filter(returned) {
				result = reflect.Append(result, returned)
			}
		}
	case reflect.Map:
		mapKeyType := containerType.Key()
		if !reflect.DeepEqual(inType, mapKeyType) {
			return nil, fmt.Errorf("mapper func arg type (%s) must be same as map key type (%s)", inType, mapKeyType)
		}
		// Iterate over the map, and call the mapper:
		for _, key := range rv.MapKeys() {
			returned := ff.Call([]reflect.Value{key})[0]
			if filter(returned) {
				result = reflect.Append(result, returned)
			}
		}
	}
	return result.Interface(), nil
}
func example_Map() {
	{
		sl := []string{"a", "b"}
		out := Map(sl, func(i int) string {
			return sl[i]
		}).([]string)
		spew.Dump(out)
	}
	{
		sl := []int{1, 2}
		out := Map(sl, func(i int) int {
			return sl[i]
		}).([]int)
		spew.Dump(out)
	}
	{
		sl := [2]int{333, 444}
		out := Map(sl, func(i int) int {
			return sl[i]
		}).([]int)
		spew.Dump(out)
	}
	{
		mp := map[string]string{
			"hello": "world",
			"foo":   "bar",
		}
		out := Map(mp, func(key string) string {
			return key
		}).([]string)
		spew.Dump(out)
	}
	{
		mp := map[string]interface{}{
			"alpha": nil,
			"beta":  "bar",
		}
		out := Map(mp, func(key string) interface{} {
			return mp[key]
		}).([]interface{})
		spew.Dump(out)
	}
}
