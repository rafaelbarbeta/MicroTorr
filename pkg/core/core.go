package core

import (
	"fmt"
	"sync"
	"time"

	"github.com/rafaelbarbeta/MicroTorr/pkg/internal"
	"github.com/rafaelbarbeta/MicroTorr/pkg/tracker"
)

// Internal Structures
type SyncPeerPieces struct {
	Piece  [][]string // PieceId as row index to PeerId who have this piece
	Speed  map[string]float64
	Rarity map[int]int // row index is the count of peers who have this piece, column index is the piece index
	Lock   sync.RWMutex
}

func InitCore(
	swarm tracker.Swarm,
	chanPeerWire chan internal.ControlMessage,
	chanCore chan internal.ControlMessage,
	chanTracker chan internal.ControlMessage,
	myId string,
	wait *sync.WaitGroup,
	verbosity int) {
	go ListenForMessages(chanPeerWire)
	for {
		time.Sleep(time.Second * 15)
		chanCore <- internal.ControlMessage{
			Opcode:  internal.HELLO,
			PeerId:  "",
			Payload: internal.HelloDebug{Msg: fmt.Sprintln("Hello from core with id: ", myId)},
		}
	}
}

func ListenForMessages(
	chanPeerWire chan internal.ControlMessage,
) {
	for {
		msg := <-chanPeerWire
		if msg.Opcode == internal.HELLO {
			fmt.Println(msg.Payload.(internal.HelloDebug).Msg)
		}
	}
}
