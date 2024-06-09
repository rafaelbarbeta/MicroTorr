package core

import (
	"crypto/sha1"
	"fmt"
	"math"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rafaelbarbeta/MicroTorr/pkg/messages"
	"github.com/rafaelbarbeta/MicroTorr/pkg/mtorr"
	"github.com/rafaelbarbeta/MicroTorr/pkg/utils"
)

const (
	OPPORTUNISTIC_CHOICE = 0.85
	WAIT_DEFAULT_TIME    = 5 * time.Second
)

func InitCore(
	mtorrent mtorr.Mtorrent,
	chanPeerWire, chanCore, chanTracker chan messages.ControlMessage,
	myId string,
	wait *sync.WaitGroup,
	seed string,
	autoSeed bool,
	waitSeeders, waitLeechers, verbosity int,
) {
	numberOfPieces := int(math.Ceil(
		float64(mtorrent.Info.Length) / float64(mtorrent.Info.Piece_length),
	))

	chanPieceRequester := make(chan messages.ControlMessage)
	chanPieceUploader := make(chan messages.ControlMessage)
	// Signal handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	PeerPieces := SyncPeerPieces{
		Have:  make(map[string][]bool),
		Speed: make(map[string]float64),
		Lock:  sync.RWMutex{},
	}

	PiecesBytes := PiecesBytes{
		Pieces: make([][]byte, numberOfPieces),
		Hash:   make([]string, numberOfPieces),
		Have:   make([]bool, numberOfPieces),
	}

	SeedMode := SeedMode{
		SeedFile: seed,
		active:   seed != "",
		auto:     autoSeed,
	}

	if SeedMode.active {
		var sha1hash strings.Builder
		var data []byte
		utils.PrintVerbose(verbosity, utils.VERBOSE, "Seed Mode active")
		utils.PrintVerbose(verbosity, utils.VERBOSE, "Opening seed file ", seed)
		// Open file and insert its pieces in the PieceBytes. Each piece has size of mtorrent.Info.Piece_length
		file, err := os.Open(SeedMode.SeedFile)
		utils.Check(err, verbosity, "Error opening seed file")
		for i := 0; i < numberOfPieces; i++ {
			data = make([]byte, mtorrent.Info.Piece_length)
			n, err := file.Read(data)
			utils.Check(err, verbosity, "Error reading seed file")
			pieceHash := fmt.Sprintf("%x", sha1.Sum(data[:n]))
			sha1hash.WriteString(pieceHash)
			PiecesBytes.AddPiece(data[:n], i)
		}
		file.Close()
		// Making sure the file pieces are correct and match the mtorrent sha1 sum
		if sha1hash.String() != mtorrent.Info.Sha1sum {
			utils.PrintVerbose(verbosity, utils.CRITICAL, "Seed file SHA1 does not match with Mtorrent SHA1. Aborting...")
			utils.PrintVerbose(verbosity, utils.DEBUG, "Seed file SHA1: ", sha1hash.String(), " \nMtorrent SHA1:", mtorrent.Info.Sha1sum)
			wait.Done()
		}
		utils.PrintVerbose(verbosity, utils.VERBOSE, "File Loaded into memory")
	}

	//Load piece hashes into memory for integrity checking
	for i := 0; i < numberOfPieces*40; i += 40 {
		PiecesBytes.Hash[i/40] = mtorrent.Info.Sha1sum[i : i+40]
	}

	go ListenForMessages(
		&PeerPieces,
		&PiecesBytes,
		&SeedMode,
		numberOfPieces,
		chanPeerWire,
		chanCore,
		chanPieceRequester,
		chanPieceUploader,
		wait,
	)

	if seed == "" {
		go PieceRequester(
			&PeerPieces,
			&PiecesBytes,
			&SeedMode,
			mtorrent,
			numberOfPieces,
			chanPieceRequester,
			chanCore,
			chanTracker,
			wait,
			waitSeeders,
			waitLeechers,
			verbosity,
		)
	}

	go PieceUploader(
		&PiecesBytes,
		mtorrent,
		chanPieceUploader,
		chanCore,
		verbosity,
	)

	go HandleSignals(
		sigs,
		chanTracker,
		wait,
		verbosity,
	)

	wait.Wait()
}

func ListenForMessages(
	PeerPieces *SyncPeerPieces,
	PiecesBytes *PiecesBytes,
	SeedMode *SeedMode,
	numberOfPieces int,
	chanPeerWire, chanCore, chanPieceRequester, chanPieceUploader chan messages.ControlMessage,
	wait *sync.WaitGroup,
) {
	for {
		msg := <-chanPeerWire
		switch msg.Opcode {
		case messages.NEW_CONNECTION:
			PeerPieces.AddPeer(msg.PeerId, numberOfPieces)
			chanCore <- messages.ControlMessage{
				Opcode: messages.BITFIELD,
				PeerId: msg.PeerId,
				Payload: messages.Bitfield{
					Bitfield: PiecesBytes.Have,
				},
			}
		case messages.DEAD_CONNECTION:
			PeerPieces.DeletePeer(msg.PeerId)
			if !SeedMode.active {
				chanPieceRequester <- msg
			}
		case messages.HAVE:
			PeerPieces.AddPiece(msg.PeerId, msg.Payload.(messages.Have).PieceIndex)
		case messages.BITFIELD:
			PeerPieces.SetBitfield(msg.PeerId, msg.Payload.(messages.Bitfield))
		case messages.REQUEST:
			chanPieceUploader <- msg
		case messages.PIECE:
			if SeedMode.active {
				panic("PIECE received when in seed mode")
			}
			chanPieceRequester <- msg
		case messages.HELLO:
			fmt.Println("Received HELLO message: ",
				msg.Payload.(messages.HelloDebug).Msg)
		default:
			panic("Unknown message type received!")
		}
	}
}
