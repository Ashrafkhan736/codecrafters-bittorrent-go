package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
)

var PEER_ID = "12345678901234567890"

func printBencode() {
	bencodedValue := os.Args[2]
	decoded, _, err := decodeBencode(bencodedValue, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	jsonOutput, _ := json.Marshal(decoded)
	fmt.Println(string(jsonOutput))
}

func getTorrentInfo() {
	filename := os.Args[2]
	metaInfo := decodeMetaInfoFile(filename)
	url, ok := metaInfo["announce"]
	if !ok {
		fmt.Println("key not found tracker")
	}
	fmt.Printf("Tracker URL: %s\n", url.(string))
	infoInterface := metaInfo["info"]
	infoDict := infoInterface.(map[string]any)

	length := decodeInt(findMapKey(infoDict, "length"))
	fmt.Printf("Length: %d\n", length)

	hash := calculateInfoHash(encodeBencode(infoDict))
	fmt.Printf("Info Hash: %s\n", hex.EncodeToString(hash))

	pieceLength := decodeInt(findMapKey(infoDict, "piece length"))
	fmt.Printf("Piece Length: %d\n", pieceLength)

	pieces := decodeByteArray(findMapKey(infoDict, "pieces"))
	fmt.Printf("Pieces Hash:\n")
	for i := 0; i < len(pieces); i += 20 {
		fmt.Printf("%x\n", pieces[i:i+20])
	}
	// for key, value := range infoDict {
	// 	switch decodedValue := value.(type) {
	// 	case []byte:
	// 		fmt.Printf("key : %v value : %v\n", key, decodedValue)
	// 	case byte:
	// 		// fmt.Printf("key : %v , value : %v\n", key, string(decodedValue))
	// 		// fmt.Printf("key : %v , value : %v\n", key, decodedValue)
	// 	case int:
	// 		if key == "length" {
	// 			fmt.Printf("Length: %d\n", decodedValue)
	// 		}
	// 		// fmt.Printf("key : %v , value : %v\n", key, decodedValue)
	// 	case string:
	// 		// fmt.Printf("key : %v , value : %v\n", key, decodedValue)
	// 	default:
	// 		// fmt.Printf("type not defined for key: %s\n", key)
	// 	}
	// }
}

type TrackerResponse struct {
	Interval int    `json:"interval"`
	Peers    string `json:"peers"`
}

func discoverPeers() {
	filename := os.Args[2]
	torrentInfo := decodeMetaInfoFile(filename)
	url := findMapKey(torrentInfo, "announce").(string)
	infoDict := findMapKey(torrentInfo, "info").(map[string]any)
	infoHash := calculateInfoHash(encodeBencode(infoDict))
	length := decodeInt(findMapKey(infoDict, "length"))

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatalf("Could not create request %s", err)
	}
	queryParams := req.URL.Query()
	queryParams.Add("info_hash", string(infoHash))
	queryParams.Add("peer_id", "12345678901234567890")
	queryParams.Add("port", "6881")
	queryParams.Add("uploaded", "0")
	queryParams.Add("downloaded", "0")
	queryParams.Add("left", strconv.Itoa(length))
	queryParams.Add("compact", "1")
	req.URL.RawQuery = queryParams.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Could not send request %s", err)
	}
	defer resp.Body.Close()
	bencodeBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Body Could not be read %s", err)
	}

	trackerResponse, _, err := decodeDictionary(string(bencodeBytes), 0)
	if err != nil {
		log.Fatalf("Could not parse bencodedValue %s", err)
	}
	peers := findMapKey(trackerResponse.(map[string]any), "peers").(string)
	peersBytes := []byte(peers)
	for i := 0; i < len(peersBytes); i += 6 {
		peer := peersBytes[i : i+6]
		ipAddress := ""
		for j, ele := range peer[:4] {
			ipAddress += fmt.Sprintf("%d", int(ele))
			if j < 3 {
				ipAddress += "."
			}
		}

		fmt.Printf("%s:%d\n", ipAddress, binary.BigEndian.Uint16(peer[4:]))
	}
}

func makeHandshake() {
	filename := os.Args[2]
	ipAddress := os.Args[3]
	torrentInfo := decodeMetaInfoFile(filename)
	infoDict := findMapKey(torrentInfo, "info").(map[string]any)
	infoHash := calculateInfoHash(encodeBencode(infoDict))
	handshakeMsg := make([]byte, 68)
	handshakeMsg[0] = 19
	handshakeMsg = append(handshakeMsg, []byte("BitTorrent protocol")...)
	handshakeMsg = append(handshakeMsg, make([]byte, 8)...)
	handshakeMsg = append(handshakeMsg, infoHash...)
	handshakeMsg = append(handshakeMsg, []byte(PEER_ID)...)
	conn, err := net.Dial("tcp", ipAddress)
	if err != nil {
		log.Fatalf("Could not connect to peer %s", err)
	}
	defer conn.Close()
	_, err = conn.Write(handshakeMsg)
	if err != nil {
		log.Fatalf("Could not write to peer %s", err)
	}
	msgFromPeer := make([]byte, 68)
	_, err = conn.Read(msgFromPeer)
	if err != nil {
		log.Fatalf("Could not read from peer %s", err)
	}
	// handshakeResponse := parseHandshakeResponse(msgFromPeer[:n])
	fmt.Printf("Peer ID: %x\n", msgFromPeer[48:])
}
