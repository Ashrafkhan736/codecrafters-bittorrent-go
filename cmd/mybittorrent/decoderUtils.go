package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"unicode"
)

// Example:
// - i52e --> 52
// - i-5e --> -5
func decodeInteger(bencodedString string, index int) (value int, nextIndex int, err error) {
	isNegative := false
	// Ignore the "i" char
	index++

	if rune(bencodedString[index]) == '-' {
		isNegative = true
		index++
	}
	var eCharIndex int
	for i := index; i < len(bencodedString); i++ {
		if rune(bencodedString[i]) == 'e' {
			eCharIndex = i
			break
		}
	}
	nextIndex = eCharIndex + 1
	value, err = strconv.Atoi(bencodedString[index:eCharIndex])
	if err == nil && isNegative {
		value *= -1
	}
	return value, nextIndex, err
}

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func decodeString(bencodedString string, index int) (value string, nextIndex int, err error) {
	var firstColonIndex int

	for i := index; i < len(bencodedString); i++ {
		if bencodedString[i] == ':' {
			firstColonIndex = i
			break
		}
	}

	lengthStr := bencodedString[index:firstColonIndex]

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", 0, err
	}

	strStartIndex := firstColonIndex + 1
	value = bencodedString[strStartIndex : strStartIndex+length]
	nextIndex = strStartIndex + length
	return value, nextIndex, err
}

// Example:
// - l5:helloi52ee -> [“hello”,52]
func decodeList(bencodedString string, index int) (any, int, error) {
	// take the first index from which we have to read the string
	// itrate over the str parse the it get the nextIndex
	// store the value and again
	list := []any{}
	index++
	elementCout := 0
	eCharIndex := 0
	for i := index; i < len(bencodedString); {
		if bencodedString[i] == 'e' {
			eCharIndex = i
			break
		}
		value, nextIndex, err := decodeBencode(bencodedString, i)
		i = nextIndex
		if err != nil {
			return list, i, err
		}
		list = append(list, value)
		elementCout++
	}
	return list[:elementCout], eCharIndex + 1, nil
}

func decodeDictionary(bencodedString string, index int) (any, int, error) {
	index++
	result := map[string]any{}
	eCharIndex := 0
	for i := index; i < len(bencodedString); {
		if bencodedString[i] == 'e' {
			eCharIndex = i
			break
		}
		// fmt.Printf("value at the start of loop %d\n", i)
		key, nextIndex, err := decodeString(bencodedString, i)
		// fmt.Printf("key %v index %d\n", key, nextIndex)
		if err != nil {
			log.Fatalf("error while decoding key %v", err)
			return result, i, err
		}
		value, nextIndex, err := decodeBencode(bencodedString, nextIndex)
		if err != nil {
			log.Fatalf("error while decoding value %v", err)
			return result, i, err
		}
		// fmt.Printf("value %v index %d\n", value, nextIndex)
		result[key] = value
		i = nextIndex
	}
	return result, eCharIndex + 1, nil
}

func decodeBencode(bencodedString string, index int) (any, int, error) {
	switch true {
	case unicode.IsDigit(rune(bencodedString[index])):
		return decodeString(bencodedString, index)
	case rune(bencodedString[index]) == 'i':
		return decodeInteger(bencodedString, index)
	case rune(bencodedString[index]) == 'l':
		return decodeList(bencodedString, index)
	case rune(bencodedString[index]) == 'd':
		return decodeDictionary(bencodedString, index)
	default:
		return "", 0, fmt.Errorf("only strings are supported at the moment")
	}
}

func decodeInt(data any) int {
	decodedValue, ok := data.(int)
	if !ok {
		log.Fatalln("Failed to convert to int")
	}
	return decodedValue
}

func findMapKey(dict map[string]any, key string) any {
	result, ok := dict[key]
	if !ok {
		fmt.Printf("key not found %s\n", key)
	}
	return result
}

func decodeByteArray(data any) []byte {
	tempString, ok := data.(string)
	if !ok {
		log.Fatalln("Failed to convert to string")
	}
	decodedValue := []byte(tempString)
	return decodedValue
}

func decodeMetaInfoFile(filename string) map[string]any {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error occured while reading file %s %v", filename, err)
	}
	decodedDict, _, err := decodeDictionary(string(data), 0)
	if err != nil {
		log.Fatalf("Error occured while decoding the string %v", err)
	}
	result, ok := decodedDict.(map[string]any)
	if !ok {
		log.Fatalln("Failed to convert infterface to dict")
	}
	// fmt.Println(result)

	return result
}
