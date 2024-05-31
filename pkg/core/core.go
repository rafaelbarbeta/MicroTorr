package core

import (
	"crypto/sha1"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/rafaelbarbeta/MicroTorr/pkg/internal"
	"github.com/rafaelbarbeta/MicroTorr/pkg/mtorr"
	"github.com/rafaelbarbeta/MicroTorr/pkg/utils"
)

func InitCore(
	mtorrent mtorr.Mtorrent,
	chanPeerWire chan internal.ControlMessage,
	chanCore chan internal.ControlMessage,
	chanTracker chan internal.ControlMessage,
	myId string,
	wait *sync.WaitGroup,
	seed string,
	autoSeed bool,
	verbosity int) {
	numberOfPieces := int(math.Ceil(
		float64(mtorrent.Info.Length) / float64(mtorrent.Info.Piece_length),
	))

	chanPieceManager := make(chan internal.ControlMessage)
	//chanCoreManager := make(chan internal.ControlMessage)

	PeerPieces := SyncPeerPieces{
		Have:  make(map[string][]bool),
		Speed: make(map[string]float64),
		Lock:  sync.RWMutex{},
	}

	PiecesBytes := PiecesBytes{
		Pieces: make([][]byte, numberOfPieces),
		Have:   make([]bool, numberOfPieces),
	}

	if seed != "" {
		var sha1hash strings.Builder
		var data []byte
		utils.PrintVerbose(verbosity, utils.VERBOSE, "Opening seed file ", seed)
		// Open file and insert its pieces in the PieceBytes. Each piece has size of mtorrent.Info.Piece_length
		file, err := os.Open(seed)
		utils.Check(err, verbosity, "Error opening seed file")
		for i := 0; i < numberOfPieces; i++ {
			data = make([]byte, mtorrent.Info.Piece_length)
			n, err := file.Read(data)
			utils.Check(err, verbosity, "Error reading seed file")
			sha1hash.WriteString(fmt.Sprintf("%x", sha1.Sum(data[:n])))
			PiecesBytes.AddPiece(data[:n], i)
		}
		file.Close()
		// Making sure the file pieces are correct and match the mtorrent sha1 sum
		if sha1hash.String() != mtorrent.Info.Sha1sum {
			utils.PrintVerbose(verbosity, utils.CRITICAL, "Seed file SHA1 does not match with Mtorrent SHA1. Aborting...")
			utils.PrintVerbose(verbosity, utils.DEBUG, "Seed file SHA1: ", sha1hash.String(), " \nMtorrent SHA1:", mtorrent.Info.Sha1sum)
			wait.Done()
			chanCore <- internal.ControlMessage{Opcode: internal.EXIT}
			wait.Wait()
		}
	}

	go ListenForMessages(&PeerPieces, numberOfPieces, chanPeerWire, wait)

	wait.Wait()
}

func ListenForMessages(
	PeerPieces *SyncPeerPieces,
	numberOfPieces int,
	chanPeerWire chan internal.ControlMessage,
	wait *sync.WaitGroup,
) {
	for {
		msg := <-chanPeerWire
		switch msg.Opcode {
		case internal.NEW_CONNECTION:
			PeerPieces.AddPeer(msg.PeerId, numberOfPieces)
		case internal.DEAD_CONNECTION:
			PeerPieces.DeletePeer(msg.PeerId)
		case internal.HAVE:
			go HandleHave(msg)
		case internal.BITFIELD:
			go HandleBitfield(msg)
		case internal.REQUEST:
			go HandleRequest(msg)
		case internal.REJECT:
			go HandleReject(msg)
		case internal.PIECE:
			go HandlePiece(msg)
		case internal.HELLO:
			fmt.Println("Received HELLO message: ",
				msg.Payload.(internal.HelloDebug).Msg)
		case internal.EXIT:
			wait.Done()
		default:
			panic("Unknown message type received!")
		}
	}
}

func PieceManager(
	PeerPieces *SyncPeerPieces,
	PiecesBytes *PiecesBytes,
	numberOfPieces int,
	chanPieceManager chan internal.ControlMessage,
	chanPeerWire chan internal.ControlMessage,
	verbosity int,
) {

}
