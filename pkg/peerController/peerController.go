package peerController

import (
	"net"
	"strconv"
	"sync"

	"github.com/rafaelbarbeta/MicroTorr/pkg/internal"
	"github.com/rafaelbarbeta/MicroTorr/pkg/tracker"
	"github.com/rafaelbarbeta/MicroTorr/pkg/utils"
)

type PeerConn struct {
	Peer map[string]net.Conn
	Lock sync.RWMutex
}

func InitPeerController(
	chanPeerMngPieceMng chan internal.ControlMessage,
	chanCoordPeerMng chan internal.ControlMessage,
	swarm tracker.Swarm,
	myId string,
	verbosity int) {
	peerConns := PeerConn{Peer: make(map[string]net.Conn), Lock: sync.RWMutex{}}
	// Connect to all peers
	for peerId, peer := range swarm.Peers {
		peerConn, err := net.Dial("tcp", peer.Ip+":"+strconv.Itoa(peer.Port))
		utils.Check(err, verbosity, "Error in connecting to peer")
		peerConns.Peer[peerId] = peerConn
	}
	// Start listener for new peers
	go ConnectNewPeers(&peerConns, swarm, myId, verbosity)

}

func ConnectNewPeers(peerConn *PeerConn, swarm tracker.Swarm, myId string, verbosity int) {
	listen, err := net.Listen("tcp", swarm.Peers[myId].Ip+":"+strconv.Itoa(swarm.Peers[myId].Port))
	utils.Check(err, verbosity, "Error in setting listener")

	for {
		newPeerConn, err := listen.Accept()
		utils.Check(err, verbosity, "Error in accepting connection with peer")
		utils.PrintVerbose(verbosity, utils.VERBOSE, "New peer connected: ", newPeerConn.RemoteAddr())
		peerConn.Lock.Lock()
		peerConn.Peer[myId] = newPeerConn
		peerConn.Lock.Unlock()
	}
}
