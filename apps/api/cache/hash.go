package cache

import (
	"encoding/base64"
	"fmt"
	"hash"
	"hash/fnv"
	"reflect"
	"sort"
)

func SimpleHash(v interface{}) string {
	h := fnv.New128a()
	hashValue(v, h)
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func hashValue(v interface{}, h hash.Hash) {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Struct:
		val := reflect.ValueOf(v)
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			f := typ.Field(i)
			field := val.Field(i)
			fieldName := f.Name
			if f.PkgPath != "" && !f.Anonymous {
				// Field is unexported, skip it
				continue
			}
			h.Write([]byte(fieldName))
			hashValue(field.Interface(), h)
		}

	case reflect.Map:
		val := reflect.ValueOf(v)
		keys := val.MapKeys()
		sort.Slice(keys, func(i, j int) bool { return SimpleHash(keys[i].Interface()) < SimpleHash(keys[j].Interface()) })

		for _, key := range keys {
			hashedKey := SimpleHash(key.Interface())
			h.Write([]byte(hashedKey))

			// handle nested maps
			if reflect.TypeOf(val.MapIndex(key).Interface()).Kind() == reflect.Map {
				nestedMap := val.MapIndex(key).Interface()
				hashValue(nestedMap, h)
			} else {
				hashedValue := SimpleHash(val.MapIndex(key).Interface())
				h.Write([]byte(hashedValue))
			}
		}

	case reflect.Slice, reflect.Array:
		val := reflect.ValueOf(v)
		elems := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			elems[i] = val.Index(i).Interface()
		}
		sort.Slice(elems, func(i, j int) bool { return SimpleHash(elems[i]) < SimpleHash(elems[j]) })

		for _, elem := range elems {
			hashValue(elem, h)
		}

	default:
		hashed := fmt.Sprintf("%v", v)
		h.Write([]byte(hashed))
	}
}
