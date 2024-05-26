package downloader

import (
	//"net/http"

	"math"
	"sync"

	"github.com/rafaelbarbeta/MicroTorr/pkg/coordinator"
	"github.com/rafaelbarbeta/MicroTorr/pkg/internal"
	"github.com/rafaelbarbeta/MicroTorr/pkg/mtorr"
	"github.com/rafaelbarbeta/MicroTorr/pkg/peerController"
	"github.com/rafaelbarbeta/MicroTorr/pkg/pieceManager"
	trackercontroller "github.com/rafaelbarbeta/MicroTorr/pkg/trackerController"
	"github.com/rafaelbarbeta/MicroTorr/pkg/utils"
)

const (
	ID_LENGTH = 20
)

func Download(mtorrent mtorr.Mtorrent, intNet, port string, verbosity int) {
	var ip string
	var err error
	var PeerPieces internal.SyncPeerPieces
	var PeerSpeeds internal.SyncPeerSpeeds
	var PieceRarity internal.SyncPieceRarity
	var wait sync.WaitGroup
	if intNet != "" {
		ip, err = utils.GetInterfaceIP(intNet)
		utils.Check(err, verbosity, "Error getting interface IP", intNet)
	} else {
		ip, err = utils.GetDefaultRouteIP()
		utils.Check(err, verbosity, "Error getting IP from default route")
	}

	peerId := utils.GenerateRandomString(ID_LENGTH)

	utils.PrintVerbose(verbosity, utils.VERBOSE, "Using IP:", ip)
	swarm := trackercontroller.GetTrackerInfo(
		mtorrent.Announce,
		peerId,
		mtorrent.Info.Id,
		ip,
		port,
		verbosity)
	utils.PrintVerbose(verbosity, utils.VERBOSE, "Swarm obtained sucessfully:", swarm)
	PeerPieces = internal.SyncPeerPieces{
		Piece: make([][]string,
			int(math.Ceil(
				float64(mtorrent.Info.Length)/float64(mtorrent.Info.Piece_length)))),
		Lock: sync.RWMutex{},
	}

	PeerSpeeds = internal.SyncPeerSpeeds{
		Speed: make(map[string]float64),
		Lock:  sync.RWMutex{},
	}

	PieceRarity = internal.SyncPieceRarity{
		Rarity: make([][]int, len(swarm.Peers)+1),
		Lock:   sync.RWMutex{},
	}
	utils.PrintVerbose(verbosity, utils.VERBOSE, "All Structures Initialized")
	chanCoordPieceMng := make(chan internal.ControlMessage)
	chanCoordPeerMng := make(chan internal.ControlMessage)
	chanPeerMngPieceMng := make(chan internal.ControlMessage)
	chanTracker := make(chan bool)

	utils.PrintVerbose(verbosity, utils.VERBOSE, "Starting all components")
	wait.Add(1)
	// Initializes all components in separated go routines
	go trackercontroller.InitTrackerController(
		mtorrent.Announce,
		peerId,
		mtorrent.Info.Id,
		ip,
		port,
		verbosity,
		chanTracker,
	)

	go peerController.InitPeerController(
		chanPeerMngPieceMng,
		chanCoordPeerMng,
		swarm,
		peerId,
		verbosity,
	)

	go coordinator.InitCoordinator(
		&PeerPieces,
		&PeerSpeeds,
		&PieceRarity,
		chanCoordPeerMng,
		chanCoordPieceMng,
		&wait,
		verbosity,
	)

	go pieceManager.InitPieceManager(
		&PeerPieces,
		&PeerSpeeds,
		&PieceRarity,
		chanCoordPieceMng,
		chanPeerMngPieceMng,
		verbosity,
	)

	wait.Wait()
}
