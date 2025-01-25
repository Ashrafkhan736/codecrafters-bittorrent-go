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

	if rune(bencodedString[0]) == '-' {
		isNegative = true
		index++
	}
	var eCharIndex int
	for i := index; i < len(bencodedString); i++ {
		if rune(bencodedString[i]) == 'e' {
			eCharIndex = i
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

func decodeBencode(bencodedString string) (interface{}, error) {
	if unicode.IsDigit(rune(bencodedString[0])) {
		value, _, err := decodeString(bencodedString, 0)
		return value, err
	} else if rune(bencodedString[0]) == 'i' {
		value, _, err := decodeInteger(bencodedString, 0)
		return value, err
	} else {
		return "", fmt.Errorf("only strings are supported at the moment")
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	command := os.Args[1]

	if command == "decode" {
		// Uncomment this block to pass the first stage
		//
		bencodedValue := os.Args[2]

		decoded, err := decodeBencode(bencodedValue)
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
