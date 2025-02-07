package protocol

import (
	"github.com/google/uuid"
	"log"
	"reflect"
	"runtime/debug"
)

func DecodeCharacterID(array []int8) uuid.UUID {
	b := make([]byte, len(array))
	for i, v := range array {
		b[i] = byte(v)
	}

	b[0], b[1], b[2], b[3] = b[3], b[2], b[1], b[0]
	b[4], b[5] = b[5], b[4]
	b[6], b[7] = b[7], b[6]

	unique, err := uuid.FromBytes(b)
	if err != nil {
		return uuid.Nil
	}

	return unique
}

func DecodeInteger(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	default:
		return 0
	}
}

func DecodeInt64(value interface{}) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int32:
		return int64(v)
	case int16:
		return int64(v)
	case int8:
		return int64(v)
	default:
		log.Printf("Unknown integer64 type: %T", value)
		return 0
	}
}

func DecodeIntegers(array interface{}) []int {
	v := reflect.ValueOf(array)
	if v.Kind() != reflect.Slice {
		log.Printf("Input is not a slice: %v", v.Type())
		debug.PrintStack()
		return nil
	}

	result := make([]int, v.Len())
	for i := 0; i < v.Len(); i++ {
		result[i] = int(v.Index(i).Int())
	}

	return result
}

func DecodeIntegers64(array interface{}) []int64 {
	v := reflect.ValueOf(array)
	if v.Kind() != reflect.Slice {
		log.Printf("Input is not a slice: %v", v.Type())
		debug.PrintStack()
		return nil
	}

	result := make([]int64, v.Len())
	for i := 0; i < v.Len(); i++ {
		result[i] = v.Index(i).Int()
	}

	return result
}

func DecodeFloats32(i interface{}) []float32 {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Slice {
		log.Printf("Input is not a slice: %v", v.Type())
		debug.PrintStack()
		return nil
	}

	result := make([]float32, v.Len())
	for i := 0; i < v.Len(); i++ {
		result[i] = float32(v.Index(i).Float())
	}

	return result
}
