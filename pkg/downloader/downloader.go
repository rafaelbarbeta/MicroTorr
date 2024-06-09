package downloader

import (
	//"net/http"

	"math/rand"
	"sync"
	"time"

	"github.com/rafaelbarbeta/MicroTorr/pkg/core"
	"github.com/rafaelbarbeta/MicroTorr/pkg/messages"
	"github.com/rafaelbarbeta/MicroTorr/pkg/mtorr"
	"github.com/rafaelbarbeta/MicroTorr/pkg/peerWire"
	trackercontroller "github.com/rafaelbarbeta/MicroTorr/pkg/trackerController"
	"github.com/rafaelbarbeta/MicroTorr/pkg/utils"
)

const (
	ID_LENGTH         = 20
	MAX_CHAN_TRACKER  = 100
	MAX_CHAN_MESSAGES = 1000
)

func Download(
	mtorrent mtorr.Mtorrent,
	intNet, port, seed string,
	autoSeed bool,
	waitSeeders, waitLeechers, maxDownSpeed, maxUpSpeed, verbosity int,
) {
	var ip string
	var err error
	var wait sync.WaitGroup

	rand.Seed(time.Now().UnixNano())

	if intNet != "" {
		ip, err = utils.GetInterfaceIP(intNet)
		utils.Check(err, verbosity, "Error getting interface IP", intNet)
	} else {
		ip, err = utils.GetDefaultRouteIP()
		utils.Check(err, verbosity, "Error getting IP from default route")
	}

	peerId := utils.GenerateRandomString(ID_LENGTH)

	utils.PrintVerbose(verbosity, utils.VERBOSE, "My Peer Id (Capped):", peerId[:5])

	utils.PrintVerbose(verbosity, utils.VERBOSE, "Using IP:", ip)
	swarm := trackercontroller.GetTrackerInfo(
		mtorrent.Announce,
		peerId,
		mtorrent.Info.Id,
		ip,
		port,
		verbosity)

	chanTracker := make(chan messages.ControlMessage)
	chanPeerWire := make(chan messages.ControlMessage, MAX_CHAN_MESSAGES)
	chanCore := make(chan messages.ControlMessage, MAX_CHAN_MESSAGES)

	utils.PrintVerbose(verbosity, utils.INFORMATION, "All Structures Initialized")
	utils.PrintVerbose(verbosity, utils.INFORMATION, "Starting components")
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

	go peerWire.InitPeerWire(
		swarm,
		peerId,
		chanPeerWire,
		chanCore,
		&wait,
		maxDownSpeed,
		maxUpSpeed,
		verbosity,
	)

	go core.InitCore(
		mtorrent,
		chanPeerWire,
		chanCore,
		chanTracker,
		peerId,
		&wait,
		seed,
		autoSeed,
		waitSeeders,
		waitLeechers,
		verbosity,
	)

	wait.Wait()
}
