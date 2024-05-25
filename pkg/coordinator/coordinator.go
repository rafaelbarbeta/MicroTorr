package coordinator

import (
	"sync"

	"github.com/rafaelbarbeta/MicroTorr/pkg/internal"
)

func InitCoordinator(
	PeerPieces *internal.SyncPeerPieces,
	PeerSpeeds *internal.SyncPeerSpeeds,
	PieceRarity *internal.SyncPieceRarity,
	chanCoordPeerMng chan internal.ControlMessage,
	chanCoordPieceMng chan internal.ControlMessage,
	wait *sync.WaitGroup) {
	//TODO
}
