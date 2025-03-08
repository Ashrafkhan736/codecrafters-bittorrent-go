package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"sort"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	command := os.Args[1]

	switch command {
	case "decode":
		printBencode()
	case "info":
		getTorrentInfo()
	case "peers":
		discoverPeers()
	case "handshake":
		makeHandshake()
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

func calculateInfoHash(bencodedString string) []byte {
	hash := sha1.New()
	hash.Write([]byte(bencodedString))
	infoHash := hash.Sum(nil)
	// Convert the hash to a hexadecimal string
	// return hex.EncodeToString(infoHash)
	return infoHash
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
