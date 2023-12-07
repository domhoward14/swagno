package utils

import (
	"crypto/sha256"
	"fmt"
	"hash/fnv"
	"reflect"
	"sort"
)

func ConvertInterfaceToMap(t interface{}) map[string]interface{} {
	m := make(map[string]interface{})

	reflectVal := reflect.ValueOf(t)
	iter := reflectVal.MapRange()
	for iter.Next() {
		m[iter.Key().String()] = iter.Value().Interface()
	}

	return m
}

func ConvertToString(t interface{}) string {
	return fmt.Sprintf("%T_%+[1]v", t)
}

func GenerateHash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return fmt.Sprint(h.Sum32())
}

func GetHashOfMap(m map[string]interface{}) string {
	h := sha256.New()

	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := m[k]
		b := sha256.Sum256([]byte(fmt.Sprintf("%v", k)))
		h.Write(b[:])
		b = sha256.Sum256([]byte(fmt.Sprintf("%v", v)))
		h.Write(b[:])
	}

	return fmt.Sprintf("%x", h.Sum(nil))[:10]
}
