package pieceManager

import (
	"fmt"
	"time"

	"github.com/rafaelbarbeta/MicroTorr/pkg/internal"
)

func InitPieceManager(
	PeerPieces *internal.SyncPeerPieces,
	PeerSpeeds *internal.SyncPeerSpeeds,
	PieceRarity *internal.SyncPieceRarity,
	chanCoordPieceMng chan internal.ControlMessage,
	chanPieceMng chan internal.ControlMessage,
	verbosity int) {
	for {
		fmt.Println("Piece Manager")
		time.Sleep(time.Second * 120)
	}
}
