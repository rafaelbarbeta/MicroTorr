package core

import (
	"crypto/sha1"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rafaelbarbeta/MicroTorr/pkg/messages"
	"github.com/rafaelbarbeta/MicroTorr/pkg/mtorr"
	"github.com/rafaelbarbeta/MicroTorr/pkg/utils"
)

const (
	OPPORTUNISTIC_CHOICE = 0.7
)

func InitCore(
	mtorrent mtorr.Mtorrent,
	chanPeerWire chan messages.ControlMessage,
	chanCore chan messages.ControlMessage,
	chanTracker chan messages.ControlMessage,
	myId string,
	wait *sync.WaitGroup,
	seed string,
	autoSeed bool,
	verbosity int) {
	numberOfPieces := int(math.Ceil(
		float64(mtorrent.Info.Length) / float64(mtorrent.Info.Piece_length),
	))

	chanPieceRequester := make(chan messages.ControlMessage)
	chanPieceUploader := make(chan messages.ControlMessage)

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
			wait,
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

	wait.Wait()
}

func ListenForMessages(
	PeerPieces *SyncPeerPieces,
	PiecesBytes *PiecesBytes,
	SeedMode *SeedMode,
	numberOfPieces int,
	chanPeerWire chan messages.ControlMessage,
	chanCore chan messages.ControlMessage,
	chanPieceRequester chan messages.ControlMessage,
	chanPieceUploader chan messages.ControlMessage,
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

func PieceRequester(
	PeerPieces *SyncPeerPieces,
	PiecesBytes *PiecesBytes,
	SeedMode *SeedMode,
	mtorrrent mtorr.Mtorrent,
	numberOfPieces int,
	chanPieceRequester chan messages.ControlMessage,
	chanCore chan messages.ControlMessage,
	wait *sync.WaitGroup,
	verbosity int,
) {
	var piecesIdx []int
	var peers [][]string
	var selectedPeer string
	var selectedPiece, selectedPieceIdx int
	var msg messages.ControlMessage
	var timeStart time.Time

	utils.PrintVerbose(verbosity, utils.INFORMATION, "Waiting for seeder..")
	// Wait until at least one seeder connects to the Swarm
	for !PeerPieces.HasSeeder() {
		utils.PrintVerbose(verbosity, utils.DEBUG, "Seeder not in swarm. Waiting")
		time.Sleep(1 * time.Second)
	}
	utils.PrintVerbose(verbosity, utils.INFORMATION, "Seeder has been added to swarm")

	for {
		piecesIdx, peers = PeerPieces.RarestPieces(PiecesBytes, numberOfPieces)
		if len(piecesIdx) == 0 { // Only happens if all pieces have been downloaded already
			AssemblePieces(mtorrrent, PiecesBytes, SeedMode, wait, verbosity)
			break
		}
		selectedPiece, selectedPieceIdx = utils.RandomChoiceInt(piecesIdx)
		// Minimum chance of chosing a random peer regardless of it being the quickest
		if utils.RandomPercentChance(OPPORTUNISTIC_CHOICE) {
			selectedPeer = PeerPieces.QuickestPeer(peers[selectedPieceIdx])
		} else {
			utils.PrintVerbose(verbosity, utils.DEBUG, "Trying a random peer instead a quick peer...")
			selectedPeer, _ = utils.RandomChoiceString(peers[selectedPieceIdx])
		}
		utils.PrintVerbose(verbosity, utils.DEBUG, "Requesting piece ", selectedPiece, " from peer ", selectedPeer)

		timeStart = time.Now() // To calculate download speed
		chanCore <- messages.ControlMessage{
			Opcode: messages.REQUEST,
			PeerId: selectedPeer,
			Payload: messages.Request{
				PieceIndex: selectedPiece,
			},
		}

		for {
			msg = <-chanPieceRequester
			if msg.PeerId != selectedPeer && msg.Opcode == messages.DEAD_CONNECTION {
				continue
			} else if msg.PeerId != selectedPeer {
				utils.PrintVerbose(verbosity, utils.CRITICAL,
					"Received unsolicited message from ",
					selectedPeer[:5])
				continue
			}
			switch msg.Opcode {
			case messages.PIECE:
				duration := time.Since(timeStart).Seconds()
				speed := float64(mtorrrent.Info.Piece_length) / duration
				go func(selectedPiece int) { //Making sure gthe hashes match. Not ideal, but errors out if they don't
					hashPiece := fmt.Sprintf("%x",
						sha1.Sum(msg.Payload.(messages.Piece).Data),
					)
					if hashPiece != PiecesBytes.Hash[selectedPiece] {
						utils.PrintVerbose(verbosity, utils.CRITICAL, "Piece Hash does not match!")
						panic("Received Piece that doesn't match hashes!")
					}
				}(selectedPiece)
				PeerPieces.SetSpeed(selectedPeer, speed)
				PiecesBytes.AddPiece(msg.Payload.(messages.Piece).Data, msg.Payload.(messages.Piece).PieceIndex)
				utils.PrintVerbose(verbosity, utils.DEBUG,
					"Piece: ",
					msg.Payload.(messages.Piece).PieceIndex,
					" from: ", selectedPeer[:5],
					"speed: ",
					fmt.Sprintf("%.2f MB/s", speed/1000000.0),
				)
			case messages.REJECT:
				PeerPieces.SetSpeed(selectedPeer, -1) // Avoids choosing this peer again
				utils.PrintVerbose(verbosity, utils.DEBUG,
					"Peer",
					selectedPeer[:5],
					"rejected piece ",
					msg.Payload.(messages.Reject).PieceIndex)
			case messages.DEAD_CONNECTION:
				utils.PrintVerbose(verbosity, utils.DEBUG,
					"Peer", selectedPeer[:5], "cannot send piece ",
					msg.Payload.(messages.Reject).PieceIndex,
					" because it is dead")
			default:
				panic("Unknown message type received at PieceRequester!")
			}
			chanCore <- messages.ControlMessage{
				Opcode: messages.HAVE,
				PeerId: "",
				Payload: messages.Have{
					PieceIndex: selectedPiece,
				},
			}
			break
		}
	}
}

func AssemblePieces(
	mtorrent mtorr.Mtorrent,
	PiecesBytes *PiecesBytes,
	SeedMode *SeedMode,
	wait *sync.WaitGroup,
	verbosity int,
) {
	utils.PrintVerbose(verbosity, utils.VERBOSE, "All pieces downloaded. Assembling...")

	var data []byte
	for i := 0; i < len(PiecesBytes.Pieces); i++ {
		data = append(data, PiecesBytes.Pieces[i]...)
	}
	fmt.Println(len(data), len(fmt.Sprintf("%x", sha1.Sum(data))))
	// Checks the Sha1sum
	if fmt.Sprintf("%x", sha1.Sum(data)) != mtorrent.Info.Id {
		utils.PrintVerbose(verbosity, utils.CRITICAL, "Assembled pieces SHA1 does not match with Mtorrent SHA1!.")
		utils.PrintVerbose(verbosity, utils.DEBUG, "Assembled pieces SHA1: ", fmt.Sprintf("%x", sha1.Sum(data)), " \nMtorrent SHA1:", mtorrent.Info.Id)
	} else {
		utils.PrintVerbose(verbosity, utils.INFORMATION, "Assembled pieces SHA1 matches with Mtorrent SHA1")
	}

	utils.PrintVerbose(verbosity, utils.INFORMATION, "Dumping Data...")
	err := os.WriteFile(mtorrent.Info.Name, data, 0644)
	utils.Check(err, verbosity, "Failed to write assembled data to disk")
	utils.PrintVerbose(verbosity, utils.VERBOSE, "Data dumped to disk")
	if SeedMode.auto {
		utils.PrintVerbose(verbosity, utils.VERBOSE, "Changed to seeding mode")
		SeedMode.active = true
		SeedMode.SeedFile = mtorrent.Info.Name
	} else {
		wait.Done()
	}
}

func PieceUploader(
	PiecesBytes *PiecesBytes,
	mtorrent mtorr.Mtorrent,
	chanPieceUploader chan messages.ControlMessage,
	chanCore chan messages.ControlMessage,
	verbosity int,
) {
	for {
		msg := <-chanPieceUploader
		go func() {
			chanCore <- messages.ControlMessage{
				Opcode: messages.PIECE,
				PeerId: msg.PeerId,
				Payload: messages.Piece{
					PieceIndex: msg.Payload.(messages.Request).PieceIndex,
					Data:       PiecesBytes.Pieces[msg.Payload.(messages.Request).PieceIndex],
				},
			}
			utils.PrintVerbose(
				verbosity, utils.DEBUG,
				"Sent piece ", msg.Payload.(messages.Request).PieceIndex,
				" to: ", msg.PeerId[:5],
			)
		}()
	}
}
