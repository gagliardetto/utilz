package utilz

import (
	"fmt"
	"reflect"
	"strings"

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
// If the mapper func has one return arg, then Map returns an array of the mapper return argument type;
// if the mapper func has two return arguments, then Map returns a map with key the first return arg, and value the second return arg.
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

var emptyBool bool
var boolType = reflect.ValueOf(emptyBool).Type()

var emptyString string
var stringType = reflect.ValueOf(emptyString).Type()

func mapWithFilter(cont interface{}, mapper interface{}, filter func(reflect.Value) bool) (interface{}, error) {
	// mapper must be a func:
	if mapperKind := reflect.TypeOf(mapper).Kind(); mapperKind != reflect.Func {
		return nil, fmt.Errorf("mapper is not a func, but a %s", mapperKind)
	}
	ff := reflect.ValueOf(mapper)
	// mapper must have one parameter:
	if ff.Type().NumIn() != 1 && ff.Type().NumIn() != 2 {
		return nil, fmt.Errorf("wrong number of parameters for mapper func: want 1 or 2, got %v", ff.Type().NumIn())
	}
	// mapper must have one return value:
	if ff.Type().NumOut() != 1 && ff.Type().NumOut() != 2 {
		return nil, fmt.Errorf("wrong number of return arguments for mapper func: want 1 or 2, got %v", ff.Type().NumOut())
	}

	containerType := reflect.TypeOf(cont)
	inType0 := ff.Type().In(0)
	outType0 := ff.Type().Out(0)

	numIn := ff.Type().NumIn()
	if numIn == 2 {
		inType1 := ff.Type().In(1)
		if !reflect.DeepEqual(inType1, containerType.Elem()) {
			return nil, fmt.Errorf("mapper func arg[1] type must be a %s, but got %s", containerType.Elem(), inType1)
		}
	}

	rv := reflect.ValueOf(cont)
	switch ff.Type().NumOut() {
	case 1:
		{
			resultSlice := reflect.MakeSlice(reflect.SliceOf(outType0), 0, 0)

			switch containerType.Kind() {
			case reflect.Slice, reflect.Array:
				if !reflect.DeepEqual(inType0, integerType) {
					return nil, fmt.Errorf("mapper func arg type must be an int, but got %s", inType0)
				}
				// Iterate over the slice/array, and call the mapper:
				for index := 0; index < rv.Len(); index++ {
					callParams := []reflect.Value{reflect.ValueOf(index)}
					if numIn == 2 {
						callParams = append(callParams, rv.Index(index))
					}
					returned := ff.Call(callParams)[0]
					if filter(returned) {
						resultSlice = reflect.Append(resultSlice, returned)
					}
				}
			case reflect.Map:
				mapKeyType := containerType.Key()
				if !reflect.DeepEqual(inType0, mapKeyType) {
					return nil, fmt.Errorf("mapper func arg type (%s) must be same as map key type (%s)", inType0, mapKeyType)
				}
				// Iterate over the map, and call the mapper:
				for _, key := range rv.MapKeys() {
					callParams := []reflect.Value{key}
					if numIn == 2 {
						callParams = append(callParams, rv.MapIndex(key))
					}
					returned := ff.Call(callParams)[0]
					if filter(returned) {
						resultSlice = reflect.Append(resultSlice, returned)
					}
				}
			}
			return resultSlice.Interface(), nil
		}
	case 2:
		{
			outType1 := ff.Type().Out(1)

			var keyType = outType0
			var valueType = outType1
			var aMapType = reflect.MapOf(keyType, valueType)
			resultMap := reflect.MakeMapWithSize(aMapType, 0)

			switch containerType.Kind() {
			case reflect.Slice, reflect.Array:
				if !reflect.DeepEqual(inType0, integerType) {
					return nil, fmt.Errorf("mapper func arg type must be an int, but got %s", inType0)
				}
				// Iterate over the slice/array, and call the mapper:
				for index := 0; index < rv.Len(); index++ {
					callParams := []reflect.Value{reflect.ValueOf(index)}
					if numIn == 2 {
						callParams = append(callParams, rv.Index(index))
					}
					returned := ff.Call(callParams)
					key, value := returned[0], returned[1]
					resultMap.SetMapIndex(key, value)
				}
			case reflect.Map:
				mapKeyType := containerType.Key()
				if !reflect.DeepEqual(inType0, mapKeyType) {
					return nil, fmt.Errorf("mapper func arg type (%s) must be same as map key type (%s)", inType0, mapKeyType)
				}
				// Iterate over the map, and call the mapper:
				for _, key := range rv.MapKeys() {
					callParams := []reflect.Value{key}
					if numIn == 2 {
						callParams = append(callParams, rv.MapIndex(key))
					}
					returned := ff.Call(callParams)
					key, value := returned[0], returned[1]
					resultMap.SetMapIndex(key, value)
				}
			}

			return resultMap.Interface(), nil
		}
	default:
		return nil, fmt.Errorf("Expected a slice/array/map, but got %s", containerType.Kind())
	}
}
func example_Map() {
	// Map to a slice:
	{
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
	// Map to a map:
	{
		{
			mp := map[string]interface{}{
				"alpha": nil,
				"beta":  "bar",
			}
			out := Map(mp, func(key string) (string, string) {
				return "hello-" + key, "world-" + key
			}).(map[string]string)
			spew.Dump(out)
		}
	}
}

// Filter filters a slice/array/map with the provided filter function.
// The filter must be a function
// with one parameter (an int for slices/arrays, or the map key type for maps),
// and one return argument (a bool).
// Filter returns the same type ast the provided value, only with elements
// included by the filter.
// NOTE: arrays become slices.
func Filter(cont interface{}, filter interface{}) interface{} {
	out, err := doFilter(cont, filter)
	if err != nil {
		panic(err)
	}
	return out
}
func doFilter(cont interface{}, filter interface{}) (interface{}, error) {
	// filter must be a func:
	if mapperKind := reflect.TypeOf(filter).Kind(); mapperKind != reflect.Func {
		return nil, fmt.Errorf("filter is not a func, but a %s", mapperKind)
	}

	ff := reflect.ValueOf(filter)
	containerType := reflect.TypeOf(cont)
	inType0 := ff.Type().In(0)
	outType := ff.Type().Out(0)

	// filter must have one parameter:
	if ff.Type().NumIn() != 1 && ff.Type().NumIn() != 2 {
		return nil, fmt.Errorf("wrong number of parameters for filter func: want 1 or 2, got %v", ff.Type().NumIn())
	}
	// filter must have one return value:
	if ff.Type().NumOut() != 1 {
		return nil, fmt.Errorf("wrong number of return arguments for filter func: want 1, got %v", ff.Type().NumOut())
	}
	// filter return type must be a bool:
	if !reflect.DeepEqual(outType, boolType) {
		return nil, fmt.Errorf("filter func return arg must be a bool, but got %s", outType)
	}

	numIn := ff.Type().NumIn()
	if numIn == 2 {
		inType1 := ff.Type().In(1)
		if !reflect.DeepEqual(inType1, containerType.Elem()) {
			return nil, fmt.Errorf("filter func arg[1] type must be a %s, but got %s", containerType.Elem(), inType1)
		}
	}

	rv := reflect.ValueOf(cont)

	switch containerType.Kind() {
	case reflect.Slice, reflect.Array:
		{
			resultSlice := reflect.MakeSlice(reflect.SliceOf(containerType.Elem()), 0, 0)
			if !reflect.DeepEqual(inType0, integerType) {
				return nil, fmt.Errorf("filter func arg type must be an int, but got %s", inType0)
			}
			// Iterate over the slice/array, and call the filter:
			for index := 0; index < rv.Len(); index++ {
				callParams := []reflect.Value{reflect.ValueOf(index)}
				if numIn == 2 {
					callParams = append(callParams, rv.Index(index))
				}
				returned := ff.Call(callParams)[0]
				if returned.Bool() {
					resultSlice = reflect.Append(resultSlice, rv.Index(index))
				}
			}
			return resultSlice.Interface(), nil
		}
	case reflect.Map:
		{
			mapKeyType := containerType.Key()
			if !reflect.DeepEqual(inType0, mapKeyType) {
				return nil, fmt.Errorf("filter func arg type (%s) must be same as map key type (%s)", inType0, mapKeyType)
			}
			resultMap := reflect.MakeMapWithSize(containerType, 0)
			// Iterate over the map, and call the filter:
			for _, key := range rv.MapKeys() {
				callParams := []reflect.Value{key}
				if numIn == 2 {
					callParams = append(callParams, rv.MapIndex(key))
				}
				returned := ff.Call(callParams)[0]
				if returned.Bool() {
					resultMap.SetMapIndex(key, rv.MapIndex(key))
				}
			}
			return resultMap.Interface(), nil
		}
	default:
		return nil, fmt.Errorf("Expected a slice/array/map, but got %s", containerType.Kind())
	}
}
func example_Filter() {
	{
		sl := []string{"a", "b"}
		out := Filter(sl, func(i int) bool {
			return sl[i] == "b"
		}).([]string)
		spew.Dump(out)
	}
	{
		sl := []int{1, 2}
		out := Filter(sl, func(i int) bool {
			return sl[i] == 1
		}).([]int)
		spew.Dump(out)
	}
	{
		sl := [2]int{333, 444}
		out := Filter(sl, func(i int) bool {
			return sl[i] == 444
		}).([]int)
		spew.Dump(out)
	}
	{
		mp := map[string]string{
			"hello": "world",
			"foo":   "bar",
		}
		out := Filter(mp, func(key string) bool {
			return key == "foo"
		}).(map[string]string)
		spew.Dump(out)
	}
	{
		mp := map[string]interface{}{
			"alpha": nil,
			"beta":  "beta-value",
			"gamma": "gamma-value",
		}
		out := Filter(mp, func(key string) bool {
			val, ok := mp[key].(string)
			if ok {
				return strings.Contains(val, "-value")
			}
			return false
		}).(map[string]interface{})
		spew.Dump(out)
	}
}

// Unique returns a deduplicated copy of the provided slice/array/map.
func Unique(cont interface{}, idGetter interface{}) interface{} {
	res, err := doUnique(cont, idGetter)
	if err != nil {
		panic(err)
	}
	return res
}
func doUnique(cont interface{}, idGetter interface{}) (interface{}, error) {
	// idGetter must be a func:
	if mapperKind := reflect.TypeOf(idGetter).Kind(); mapperKind != reflect.Func {
		return nil, fmt.Errorf("idGetter is not a func, but a %s", mapperKind)
	}

	ff := reflect.ValueOf(idGetter)
	containerType := reflect.TypeOf(cont)
	inType0 := ff.Type().In(0)
	outType := ff.Type().Out(0)

	// idGetter must have one parameter:
	if ff.Type().NumIn() != 1 && ff.Type().NumIn() != 2 {
		return nil, fmt.Errorf("wrong number of parameters for idGetter func: want 1 or 2, got %v", ff.Type().NumIn())
	}
	// idGetter must have one return value:
	if ff.Type().NumOut() != 1 {
		return nil, fmt.Errorf("wrong number of return arguments for idGetter func: want 1, got %v", ff.Type().NumOut())
	}
	// idGetter return type must be a string:
	if !reflect.DeepEqual(outType, stringType) {
		return nil, fmt.Errorf("idGetter func return arg must be a string, but got %s", outType)
	}

	numIn := ff.Type().NumIn()
	if numIn == 2 {
		inType1 := ff.Type().In(1)
		if !reflect.DeepEqual(inType1, containerType.Elem()) {
			return nil, fmt.Errorf("idGetter func arg[1] type must be a %s, but got %s", containerType.Elem(), inType1)
		}
	}

	rv := reflect.ValueOf(cont)
	storeIndex := hashsearch.New()

	switch containerType.Kind() {
	case reflect.Slice, reflect.Array:
		{
			resultSlice := reflect.MakeSlice(reflect.SliceOf(containerType.Elem()), 0, 0)
			if !reflect.DeepEqual(inType0, integerType) {
				return nil, fmt.Errorf("idGetter func arg type must be an int, but got %s", inType0)
			}
			// Iterate over the slice/array, and call the idGetter:
			for index := 0; index < rv.Len(); index++ {
				callParams := []reflect.Value{reflect.ValueOf(index)}
				if numIn == 2 {
					callParams = append(callParams, rv.Index(index))
				}
				id := ff.Call(callParams)[0]
				if !storeIndex.Has(id.String()) {
					resultSlice = reflect.Append(resultSlice, rv.Index(index))
					storeIndex.OrderedAppend(id.String())
				}
			}
			return resultSlice.Interface(), nil
		}
	case reflect.Map:
		{
			mapKeyType := containerType.Key()
			if !reflect.DeepEqual(inType0, mapKeyType) {
				return nil, fmt.Errorf("idGetter func arg type (%s) must be same as map key type (%s)", inType0, mapKeyType)
			}
			resultMap := reflect.MakeMapWithSize(containerType, 0)
			// Iterate over the map, and call the idGetter:
			for _, key := range rv.MapKeys() {
				callParams := []reflect.Value{key}
				if numIn == 2 {
					callParams = append(callParams, rv.MapIndex(key))
				}
				id := ff.Call(callParams)[0]
				if !storeIndex.Has(id.String()) {
					resultMap.SetMapIndex(key, rv.MapIndex(key))
					storeIndex.OrderedAppend(id.String())
				}
			}
			return resultMap.Interface(), nil
		}
	default:
		return nil, fmt.Errorf("Expected a slice/array/map, but got %s", containerType.Kind())
	}
}
func example_Unique() {
	{
		sl := []string{
			"a",
			"b",
			"b",
			"b",
			"b",
			"c",
			"c",
			"c",
			"c",
			"d",
			"d",
			"d",
			"d",
			"c",
			"c",
			"c",
			"c",
			"b",
			"b",
			"b",
		}
		out := Unique(sl, func(i int) string {
			return sl[i]
		}).([]string)
		spew.Dump(out)
	}
	{
		mp := map[string]string{
			"hello":   "world",
			"foo":     "bar",
			"foo-dup": "bar",
		}
		out := Unique(mp, func(key string, val string) string {
			return val
		}).(map[string]string)
		spew.Dump(out)
	}
}

type MR struct {
	v interface{}
}

func NewMR(v interface{}) *MR {
	return &MR{
		v: v,
	}
}

//
func (mr *MR) Map(mapper interface{}) *MR {
	return &MR{
		v: Map(mr.v, mapper),
	}
}

//
func (mr *MR) Filter(filter interface{}) *MR {
	return &MR{
		v: Filter(mr.v, filter),
	}
}

//
func (mr *MR) Unique(idGetter interface{}) *MR {
	return &MR{
		v: Unique(mr.v, idGetter),
	}
}

//
func (mr *MR) Out() interface{} {
	return mr.v
}

func example_MR() {
	{
		sl := []string{"a", "b", "a", "c", "b", "b", "b"}
		out := NewMR(sl).
			Map(func(i int, v string) string {
				return v
			}).
			Filter(func(i int, v string) bool {
				return v > "a"
			}).
			Map(func(i int, v string) string {
				return v + "+"
			}).
			Unique(func(key int, v string) string {
				return v + "+"
			}).
			Out().([]string)
		spew.Dump(out)
	}
}
