package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
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

func decodeDictionary(bencodedString string, index int) (interface{}, int, error) {
	index++
	result := map[string]interface{}{}
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

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	command := os.Args[1]

	switch command {
	case "decode":
		bencodedValue := os.Args[2]
		decoded, _, err := decodeBencode(bencodedValue, 0)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))

	case "info":
		filename := os.Args[2]
		metaInfo := decodeMetaInfoFile(filename)
		url, ok := metaInfo["announce"]
		if !ok {
			fmt.Println("key not found tracker")
		}
		fmt.Printf("Tracker URL: %s\n", url.(string))
		infoInterface := metaInfo["info"]
		infoDict := infoInterface.(map[string]any)
		for key, value := range infoDict {
			switch decodedValue := value.(type) {
			case []byte:
				fmt.Printf("key : %v value : %v\n", key, decodedValue)
			case byte:
				// fmt.Printf("key : %v , value : %v\n", key, string(decodedValue))
				// fmt.Printf("key : %v , value : %v\n", key, decodedValue)
			case int:
				if key == "length" {
					fmt.Printf("Length: %d\n", decodedValue)
				}
				// fmt.Printf("key : %v , value : %v\n", key, decodedValue)
			case string:
				// fmt.Printf("key : %v , value : %v\n", key, decodedValue)
			default:
				// fmt.Printf("type not defined for key: %s\n", key)
			}
		}
		hash := calculateInfoHash(encodeBencode(infoDict))
		fmt.Printf("Info Hash: %s", hash)

	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

func calculateInfoHash(bencodedString string) string {
	hash := sha1.New()
	hash.Write([]byte(bencodedString))
	infoHash := hash.Sum(nil)
	// Convert the hash to a hexadecimal string
	return hex.EncodeToString(infoHash)
}

func encodeBencode(data any) string {
	result := ""
	switch v := data.(type) {
	case int:
		result = fmt.Sprintf("i%de", v)
	case string:
		result = fmt.Sprintf("%d:%s", len(v), v)
	case []any:
		result += "l"
		for _, elem := range v {
			result += encodeBencode(elem)
		}
		result += "e"
	case map[string]any:
		result += "d"
		sortedKeys := make([]string, 0, len(v))
		for key := range v {
			sortedKeys = append(sortedKeys, key)
		}
		sort.Strings(sortedKeys)
		for _, key := range sortedKeys {
			encodedKey := encodeBencode(key)
			value, ok := v[key]
			if !ok {
				log.Fatalf("Key not found %s\n", key)
			}
			encodedValue := encodeBencode(value)
			result = result + encodedKey + encodedValue
		}
		result += "e"
	default:
		// 	fmt.Printf("%v", v)
		log.Fatalf("Undefined type %T", v)
	}
	return result
}
