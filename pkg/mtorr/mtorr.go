package mtorr

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"
	"strings"

	"github.com/jackpal/bencode-go"
)

const ()

type Mtorrent struct {
	Announce string
	Info     Info
}

type Info struct {
	Length       int
	Name         string
	Piece_length int
	Sha1sum      string
	Id           string
}

func check(e error, message ...string) {
	if e != nil {
		fmt.Println(strings.Join(message, " "))
		panic(e)
	}
}

func printVerbose(verbose bool, message ...interface{}) {
	if verbose {
		var sb strings.Builder
		sb.WriteString("[INFO]: ")
		for _, m := range message {
			sb.WriteString(fmt.Sprintf("%v", m))
		}
		fmt.Println(sb.String())
	}
}

func GenMtorrent(fileName string, tracker string, pieceLength int, verbose bool) {
	var sha1hash strings.Builder
	var bencodeBuffer bytes.Buffer
	mtorrent := Mtorrent{}

	printVerbose(verbose, "Reading file", fileName)
	data, err := os.ReadFile(fileName)
	check(err, "Error reading file")
	length := len(data)
	printVerbose(verbose, "File length:", length)

	mtorrent.Announce = tracker
	mtorrent.Info.Length = length
	mtorrent.Info.Name = fileName
	mtorrent.Info.Piece_length = pieceLength

	for i := 0; i < length; i += pieceLength {
		printVerbose(verbose, "Piece", i/pieceLength)
		sha1hash.WriteString(fmt.Sprintf("%x", sha1.Sum(data[i:i+pieceLength])))
	}

	mtorrent.Info.Sha1sum = sha1hash.String()
	mtorrent.Info.Id = fmt.Sprintf("%x", sha1.Sum(data))

	// Bencode the Mtorrent
	err = bencode.Marshal(&bencodeBuffer, mtorrent)
	check(err, "Error bencoding Mtorrent")

	err = os.WriteFile(fileName+".mtorrent", bencodeBuffer.Bytes(), 0644)
	check(err, "Error writing Mtorrent")
}

func LoadMtorrent(fileName string) Mtorrent {
	mtorrent := Mtorrent{}
	file, err := os.Open(fileName)
	check(err, "Error opening Mtorrent", fileName)
	err = bencode.Unmarshal(file, &mtorrent)
	check(err, "Error unmarshalling Mtorrent", fileName)
	file.Close()

	return mtorrent
}
