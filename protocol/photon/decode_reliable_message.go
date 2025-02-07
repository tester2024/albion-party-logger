package photon

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	NilType               = 42
	DictionaryType        = 68
	StringSliceType       = 97
	Int8Type              = 98
	Custom                = 99
	DoubleType            = 100
	EventDateType         = 101
	Float32Type           = 102
	Hashtable             = 104
	Int32Type             = 105
	Int16Type             = 107
	Int64Type             = 108
	Int32SliceType        = 110
	BooleanType           = 111
	OperationResponseType = 112
	OperationRequestType  = 113
	StringType            = 115
	Int8SliceType         = 120
	SliceType             = 121
	ObjectSliceType       = 122
)

type ReliableMessageParamaters map[uint8]interface{}

// DecodeReliableMessage Converts the parameters of a reliable message into a hash suitable for use in
func DecodeReliableMessage(msg ReliableMessage) (ReliableMessageParamaters, error) {
	buf := bytes.NewBuffer(msg.Data)
	params := make(map[uint8]interface{})

	for i := 0; i < int(msg.ParamaterCount); i++ {
		var paramID uint8
		var paramType uint8

		if err := binary.Read(buf, binary.BigEndian, &paramID); err != nil {
			return nil, err
		}

		if err := binary.Read(buf, binary.BigEndian, &paramType); err != nil {
			return nil, err
		}

		paramsKey := paramID
		decoded, err := decodeType(buf, paramType)
		if err != nil {
			return nil, err
		}
		params[paramsKey] = decoded
	}

	return params, nil
}

func decodeType(buf *bytes.Buffer, paramType uint8) (interface{}, error) {
	switch paramType {
	case NilType, 0:
		// Do nothing
		return nil, nil
	case Int8Type:
		return decodeInt8Type(buf) // buffer
	case Float32Type:
		return decodeFloat32Type(buf) // buffer
	case Int32Type:
		return decodeInt32Type(buf) // buffer
	case Int16Type, 7:
		return decodeInt16Type(buf) // buffer
	case Int64Type:
		return decodeInt64Type(buf) // buffer
	case StringType:
		return decodeStringType(buf)
	case BooleanType:
		result, err := decodeBooleanType(buf)

		if err != nil {
			return nil, fmt.Errorf("ERROR - Boolean - %v", err.Error())
		} else {
			return result, nil
		}
	case Int8SliceType:
		result, err := decodeSliceInt8Type(buf)
		if err != nil {
			return nil, fmt.Errorf("ERROR - Slice Int8 - %v", err.Error())
		} else {
			return result, nil
		}
	case SliceType:
		array, err := decodeSlice(buf)
		if err != nil {
			return nil, fmt.Errorf("ERROR - Slice - %v", err.Error())
		} else {
			return array, nil
		}
	case DictionaryType:
		dict, err := decodeDictionaryType(buf)
		if err != nil {
			return nil, fmt.Errorf("ERROR - Dictionary - %v", err.Error())
		} else {
			return dict, nil
		}
	default:
		return nil, fmt.Errorf("ERROR - Invalid type of %v", paramType)
	}
}

func decodeSlice(buf *bytes.Buffer) (interface{}, error) {
	var length uint16
	var sliceType uint8

	if err := binary.Read(buf, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	if err := binary.Read(buf, binary.BigEndian, &sliceType); err != nil {
		return nil, err
	}

	switch sliceType {
	case Float32Type:
		array := make([]float32, length)

		for j := 0; j < int(length); j++ {
			temp, err := decodeFloat32Type(buf)
			if err != nil {
				return nil, err
			}
			array[j] = temp
		}

		return array, nil
	case Int32Type:
		array := make([]int32, length)

		for j := 0; j < int(length); j++ {
			temp, err := decodeInt32Type(buf)
			if err != nil {
				return nil, err
			}
			array[j] = temp
		}

		return array, nil
	case Int16Type:
		array := make([]int16, length)

		for j := 0; j < int(length); j++ {
			var temp int16
			if err := binary.Read(buf, binary.BigEndian, &temp); err != nil {
				return nil, err
			}
			array[j] = temp
		}

		return array, nil
	case Int64Type:
		array := make([]int64, length)

		for j := 0; j < int(length); j++ {
			temp, err := decodeInt64Type(buf)
			if err != nil {
				return nil, err
			}
			array[j] = temp
		}

		return array, nil
	case StringType:
		array := make([]string, length)

		for j := 0; j < int(length); j++ {
			stringType, err := decodeStringType(buf)
			if err != nil {
				return nil, err
			}
			array[j] = stringType
		}

		return array, nil
	case BooleanType:
		array := make([]bool, length)

		for j := 0; j < int(length); j++ {
			result, err := decodeBooleanType(buf)

			if err != nil {
				return array, err
			}

			array[j] = result
		}

		return array, nil
	case Int8SliceType:
		array := make([][]int8, length)

		for j := 0; j < int(length); j++ {
			result, err := decodeSliceInt8Type(buf)
			if err != nil {
				return nil, err
			}
			array[j] = result
		}

		return array, nil
	case SliceType:
		array := make([]interface{}, length)

		for j := 0; j < int(length); j++ {
			subArray, err := decodeSlice(buf)

			if err != nil {
				return nil, err
			}

			array[j] = subArray
		}

		return array, nil
	default:
		return nil, fmt.Errorf("invalid slice type of %d", sliceType)
	}
}

func decodeInt8Type(buf *bytes.Buffer) (temp int8, err error) {
	err = binary.Read(buf, binary.BigEndian, &temp)
	return
}

func decodeFloat32Type(buf *bytes.Buffer) (temp float32, err error) {
	err = binary.Read(buf, binary.BigEndian, &temp)
	return
}

func decodeInt16Type(buf *bytes.Buffer) (temp int16, err error) {
	err = binary.Read(buf, binary.BigEndian, &temp)
	return
}

func decodeInt32Type(buf *bytes.Buffer) (temp int32, err error) {
	err = binary.Read(buf, binary.BigEndian, &temp)
	return
}

func decodeInt64Type(buf *bytes.Buffer) (temp int64, err error) {
	err = binary.Read(buf, binary.BigEndian, &temp)
	return
}

func decodeStringType(buf *bytes.Buffer) (string, error) {
	var length uint16

	err := binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		return "", err
	}

	strBytes := make([]byte, length)
	_, err = buf.Read(strBytes)
	if err != nil {
		return "", err
	}

	return string(strBytes[:]), nil
}

func decodeBooleanType(buf *bytes.Buffer) (bool, error) {
	var value uint8

	if err := binary.Read(buf, binary.BigEndian, &value); err != nil {
		return false, err
	}

	if value == 0 {
		return false, nil
	} else if value == 1 {
		return true, nil
	} else {
		return false, fmt.Errorf("invalid value for boolean of %d", value)
	}

}

func decodeSliceInt8Type(buf *bytes.Buffer) ([]int8, error) {
	var length uint32

	err := binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}

	array := make([]int8, length)

	err = binary.Read(buf, binary.BigEndian, array)
	if err != nil {
		return nil, err
	}

	return array, nil
}

func decodeDictionaryType(buf *bytes.Buffer) (map[interface{}]interface{}, error) {
	var keyTypeCode uint8
	var valueTypeCode uint8
	var dictionarySize uint16

	// Read the key type code
	err := binary.Read(buf, binary.BigEndian, &keyTypeCode)
	if err != nil {
		return nil, err
	}

	// Read the value type code
	err = binary.Read(buf, binary.BigEndian, &valueTypeCode)
	if err != nil {
		return nil, err
	}

	// Read the dictionary size
	err = binary.Read(buf, binary.BigEndian, &dictionarySize)
	if err != nil {
		return nil, err
	}

	// Initialize the dictionary map
	dictionary := make(map[interface{}]interface{})

	// Loop through the dictionary size
	for i := uint16(0); i < dictionarySize; i++ {
		// Handle key type code
		actualKeyTypeCode := keyTypeCode
		if keyTypeCode == 0 || keyTypeCode == 42 {
			err = binary.Read(buf, binary.BigEndian, &actualKeyTypeCode)
			if err != nil {
				return nil, err
			}
		}
		// Deserialize the key
		key, err := decodeType(buf, actualKeyTypeCode)
		if err != nil {
			return nil, err
		}

		// Handle value type code
		actualValueTypeCode := valueTypeCode
		if valueTypeCode == 0 || valueTypeCode == 42 {
			err = binary.Read(buf, binary.BigEndian, &actualValueTypeCode)
			if err != nil {
				return nil, err
			}
		}
		// Deserialize the value
		value, err := decodeType(buf, actualValueTypeCode)
		if err != nil {
			return nil, err
		}

		// Enchant the key-value pair to the dictionary
		dictionary[key] = value
	}

	return dictionary, nil
}
