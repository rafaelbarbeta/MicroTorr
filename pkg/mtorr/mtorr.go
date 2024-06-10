package mtorr

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"
	"strings"

	"github.com/jackpal/bencode-go"
	"github.com/rafaelbarbeta/MicroTorr/pkg/utils"
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

func GenMtorrent(fileName string, tracker string, pieceLength int, verbose int) {
	var sha1hash strings.Builder
	var bencodeBuffer bytes.Buffer
	mtorrent := Mtorrent{}

	utils.PrintVerbose(verbose, utils.DEBUG, "Reading file", fileName)
	data, err := os.ReadFile(fileName)
	utils.Check(err, verbose, "Error reading file")
	length := len(data)
	utils.PrintVerbose(verbose, utils.INFORMATION, "File length:", length)

	mtorrent.Announce = tracker
	mtorrent.Info.Length = length
	mtorrent.Info.Name = fileName
	mtorrent.Info.Piece_length = pieceLength

	for i := 0; i < length; i += pieceLength {
		p := data[i:Min(i+pieceLength, length)]
		utils.PrintVerbose(verbose, utils.DEBUG, "Piece", i/pieceLength)
		sha1hash.WriteString(fmt.Sprintf("%x", sha1.Sum(p)))
	}

	mtorrent.Info.Sha1sum = sha1hash.String()
	mtorrent.Info.Id = fmt.Sprintf("%x", sha1.Sum(data))
	utils.PrintVerbose(verbose, utils.VERBOSE, "Mtorrent:", mtorrent)

	// Bencode the Mtorrent
	err = bencode.Marshal(&bencodeBuffer, mtorrent)
	utils.Check(err, verbose, "Error bencoding Mtorrent")

	err = os.WriteFile(fileName+".mtorrent", bencodeBuffer.Bytes(), 0644)
	utils.Check(err, verbose, "Error writing Mtorrent")
}

func LoadMtorrent(fileName string, verbosity int) Mtorrent {
	mtorrent := Mtorrent{}
	file, err := os.Open(fileName)
	utils.Check(err, verbosity, "Error opening Mtorrent", fileName)
	err = bencode.Unmarshal(file, &mtorrent)
	utils.Check(err, verbosity, "Error unmarshalling Mtorrent", fileName)
	file.Close()

	return mtorrent
}

func (mtorrent Mtorrent) String() string {
	var mtorrentString string
	mtorrentString += fmt.Sprintln("Tracker Link:", mtorrent.Announce)
	mtorrentString += fmt.Sprintln("File Name:", mtorrent.Info.Name)
	mtorrentString += fmt.Sprintln("File Length:", mtorrent.Info.Length)
	mtorrentString += fmt.Sprintln("Piece Length:", mtorrent.Info.Piece_length)
	mtorrentString += fmt.Sprintln("Sha1sum (first 20 bytes):", mtorrent.Info.Sha1sum[:20])
	mtorrentString += fmt.Sprint("Id Hash:", mtorrent.Info.Id)
	return mtorrentString
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
