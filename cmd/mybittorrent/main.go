package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
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
func decodeList(bencodedString string, index int) (interface{}, int, error) {
	// take the first index from which we have to read the string
	// itrate over the str parse the it get the nextIndex
	// store the value and again
	list := []interface{}{}
	index++
	elementCout := 0
	eCharIndex := 0
	// value, index, err := decodeBencode(bencodedString, index)
	// fmt.Printf("%s %d", value, index)
	// if err != nil {
	// 	return list, index, err
	// }
	// list = append(list, value)
	for i := index; i < len(bencodedString); {
		if bencodedString[i] == 'e' {
			eCharIndex = i
			break
		}
		value, nextIndex, err := decodeBencode(bencodedString, i)
		i = nextIndex
		// fmt.Printf("%v %d\n", value, i)
		if err != nil {
			return list, i, err
		}
		list = append(list, value)
		elementCout++
	}
	return list[:elementCout], eCharIndex + 1, nil
}

func decodeBencode(bencodedString string, index int) (interface{}, int, error) {
	switch true {
	case unicode.IsDigit(rune(bencodedString[index])):
		return decodeString(bencodedString, index)
	case rune(bencodedString[index]) == 'i':
		return decodeInteger(bencodedString, index)
	case rune(bencodedString[index]) == 'l':
		return decodeList(bencodedString, index)
	default:
		return "", 0, fmt.Errorf("only strings are supported at the moment")
	}
	// if unicode.IsDigit(rune(bencodedString[0])) {
	// 	value, _, err := decodeString(bencodedString, 0)
	// 	return value, err
	// } else if rune(bencodedString[0]) == 'i' {
	// 	value, _, err := decodeInteger(bencodedString, 0)
	// 	return value, err
	// } else {
	// 	return "", fmt.Errorf("only strings are supported at the moment")
	// }
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	command := os.Args[1]

	if command == "decode" {
		// Uncomment this block to pass the first stage
		//
		bencodedValue := os.Args[2]

		decoded, _, err := decodeBencode(bencodedValue, 0)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
