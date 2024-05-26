package coordinator

import (
	"fmt"
	"sync"
	"time"

	"github.com/rafaelbarbeta/MicroTorr/pkg/internal"
)

func InitCoordinator(
	PeerPieces *internal.SyncPeerPieces,
	PeerSpeeds *internal.SyncPeerSpeeds,
	PieceRarity *internal.SyncPieceRarity,
	chanCoordPeerMng chan internal.ControlMessage,
	chanCoordPieceMng chan internal.ControlMessage,
	wait *sync.WaitGroup,
	verbosity int) {
	for {
		fmt.Println("Piece Manager")
		time.Sleep(time.Second * 120)
		wait.Done()
	}
}
