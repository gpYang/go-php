package php

import (
	"reflect"
	"sort"
)

// ArrayKeys array_keys returns the keys, numeric and string, from the array.
func ArrayKeys(array interface{}) interface{} {
	t, v, l := getCommon(array)
	res := make([]interface{}, l)
	switch t.Kind() {
	case reflect.Slice:
		for i := 0; i < l; i++ {
			res[i] = i
		}
	case reflect.Map:
		for i, k := range v.MapKeys() {
			res[i] = k.Interface()
		}
	default:
		panic("expects parameter 1 to be array")
	}
	return res
}

// ArrayValues array_values returns all the values from the array and indexes the array numerically.
func ArrayValues(array interface{}) interface{} {
	t, v, l := getCommon(array)
	res := make([]interface{}, l)
	switch t.Kind() {
	case reflect.Slice:
		for i := 0; i < l; i++ {
			res[i] = v.Index(i).Interface()
		}
	case reflect.Map:
		for i, k := range v.MapKeys() {
			res[i] = v.MapIndex(k)
		}
	default:
		panic("expects parameter 1 to be array")
	}
	return res
}

// ArrayKeyExists array_key_exists — Checks if the given key or index exists in the array
func ArrayKeyExists(key, array interface{}) bool {
	t, v, l := getCommon(array)
	switch t.Kind() {
	case reflect.Slice:
		for i := 0; i < l; i++ {
			if reflect.DeepEqual(key, i) {
				return true
			}
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			if reflect.DeepEqual(key, k.Interface()) {
				return true
			}
		}
	default:
		panic("expects parameter 2 to be array")
	}
	return false
}

// InArray in_array — Checks if a value exists in an array
func InArray(needle, haystack interface{}) bool {
	t, v, l := getCommon(haystack)
	switch t.Kind() {
	case reflect.Slice:
		for i := 0; i < l; i++ {
			if reflect.DeepEqual(needle, v.Index(i).Interface()) {
				return true
			}
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			if reflect.DeepEqual(needle, v.MapIndex(k).Interface()) {
				return true
			}
		}
	default:
		panic("expects parameter 2 to be array")
	}
	return false
}

// ArrayFilp array_flip — Exchanges all keys with their associated values in an array
func ArrayFilp(array interface{}) map[interface{}]interface{} {
	t, v, l := getCommon(array)
	res := make(map[interface{}]interface{}, l)
	switch t.Kind() {
	case reflect.Slice:
		for i := 0; i < l; i++ {
			res[v.Index(i).Interface()] = i
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			res[v.MapIndex(k).Interface()] = k.Interface()
		}
	default:
		panic("expects parameter 1 to be array")
	}
	return res
}

// ArrayUnique array_unique — Removes duplicate values from an array
func ArrayUnique(array interface{}) interface{} {
	t, v, l := getCommon(array)
	res := make(map[interface{}]int)
	switch t.Kind() {
	case reflect.Slice:
		for i := 0; i < l; i++ {
			res[v.Index(i).Interface()] = 1
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			res[v.MapIndex(k).Interface()] = 1
		}
	default:
		panic("expects parameter 1 to be array")
	}
	return ArrayKeys(res)
}

// Sort can only sort []int, []string, []float64
func Sort(array interface{}) {
	t, v, _ := getCommon(array)
	// res := make([]interface{}, l)
	if t.Kind() == reflect.Slice {
		switch v.Index(0).Kind() {
		case reflect.Int:
			array := array.([]int)
			sort.Ints(array)
		case reflect.String:
			array := array.([]string)
			sort.Strings(array)
		case reflect.Float64:
			array := array.([]float64)
			sort.Float64s(array)
		default:
			panic("the param can only be int/string/float64 array")
		}
	} else {
		panic("expects parameter 1 to be array")
	}
}

func getCommon(array interface{}) (reflect.Type, reflect.Value, int) {
	t := reflect.TypeOf(array)
	v := reflect.ValueOf(array)
	l := v.Len()
	return t, v, l
}
