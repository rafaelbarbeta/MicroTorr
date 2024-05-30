package peerWire

import (
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"

	"github.com/rafaelbarbeta/MicroTorr/pkg/internal"
	"github.com/rafaelbarbeta/MicroTorr/pkg/tracker"
	"github.com/rafaelbarbeta/MicroTorr/pkg/utils"
)

type peerConn struct {
	conns   map[string]net.Conn
	send    map[string]*gob.Encoder
	receive map[string]*gob.Decoder
	lock    sync.RWMutex
}

func InitPeerWire(
	swarm tracker.Swarm,
	myId string,
	chanPeerWire chan internal.ControlMessage,
	chanCore chan internal.ControlMessage,
	wait *sync.WaitGroup,
	verbosity int,
) {
	gob.Register(internal.HandShake{})
	gob.Register(internal.Have{})
	gob.Register(internal.Bitfield{})
	gob.Register(internal.Request{})
	gob.Register(internal.Reject{})
	gob.Register(internal.Piece{})
	gob.Register(internal.HelloDebug{})

	peerConn := peerConn{
		conns:   make(map[string]net.Conn),
		send:    make(map[string]*gob.Encoder),
		receive: make(map[string]*gob.Decoder),
		lock:    sync.RWMutex{},
	}

	// Connect to all Peers and insert than in the map
	// Also performs Handshake with each, so they know 'myId'
	for _, peer := range swarm.Peers {
		if peer.Id == myId {
			continue
		}
		utils.PrintVerbose(verbosity, utils.INFORMATION, "Connecting to peer: ", peer.Id[:5])
		conn, err := net.Dial("tcp", peer.Ip+":"+strconv.Itoa(peer.Port))
		utils.Check(err, verbosity, "Error connecting to peer: ", peer.Id[:5])
		peerConn.conns[peer.Id] = conn
		peerConn.send[peer.Id] = gob.NewEncoder(conn)
		peerConn.receive[peer.Id] = gob.NewDecoder(conn)
		_, err = PerfomHandshake(peerConn.send[peer.Id],
			peerConn.receive[peer.Id],
			myId,
			swarm.IdHash,
			verbosity)
		utils.Check(err, verbosity, "Error performing Handshake with peer: ", peer.Id[:5])
		utils.PrintVerbose(verbosity, utils.VERBOSE, "Handshake with peer: ", peer.Id[:5], "sucessful")
		go ListenForMessages(&peerConn, peer.Id, chanPeerWire, verbosity)
	}

	// Start listener for new connections
	go ListenForConns(
		&peerConn,
		myId,
		swarm.IdHash,
		swarm.Peers[myId].Ip+":"+strconv.Itoa(swarm.Peers[myId].Port),
		chanPeerWire,
		verbosity,
	)

	// Listen for Core messages to be sent out to peers
	// TODO: ADD A FOR LOOP TO LISTEN FOR MUTIPLE CORE MESSAGES
	go ListenForCoreMessages(
		&peerConn,
		chanCore,
		wait,
		verbosity,
	)

	wait.Wait()
}

func ListenForMessages(
	peerConn *peerConn,
	peerId string,
	chanPeerWire chan internal.ControlMessage,
	verbosity int,
) {
	utils.PrintVerbose(verbosity, utils.DEBUG, "Starting listening for messages from peer: ", peerId[:5])
	var msg internal.Message
	for {
		err := peerConn.receive[peerId].Decode(&msg)
		// verificar desconexão
		if err == io.EOF {
			utils.PrintVerbose(verbosity, utils.CRITICAL, "Peer: ", peerId[:5], "disconnected!")
			peerConn.lock.Lock()
			delete(peerConn.conns, peerId)
			delete(peerConn.send, peerId)
			delete(peerConn.receive, peerId)
			chanPeerWire <- internal.ControlMessage{
				Opcode:  internal.DEAD_CONNECTION,
				PeerId:  peerId,
				Payload: nil,
			}
			peerConn.lock.Unlock()
			return
		}
		utils.Check(err, verbosity, "Error receiving message from peer", peerId[:5])
		opcode := MessageOpcode(msg)
		utils.PrintVerbose(verbosity, utils.DEBUG, "Received message from peer: ", peerId[:5], "with opcode: ", opcode)
		chanPeerWire <- internal.ControlMessage{
			Opcode:  opcode,
			PeerId:  peerId,
			Payload: msg.Data,
		}
	}
}

func ListenForCoreMessages(
	peerConn *peerConn,
	chanCore chan internal.ControlMessage,
	wait *sync.WaitGroup,
	verbosity int,
) {
	var controlMsg internal.ControlMessage
	var peerMsg internal.Message
	for {
		controlMsg = <-chanCore
		// Checks if is Done already
		if controlMsg.Opcode == internal.EXIT {
			wait.Done()
			return
		}
		peerMsg = internal.Message{Data: controlMsg.Payload}
		if controlMsg.PeerId == "" { // Empty string is used to broadcast message
			peerConn.lock.Lock()
			for _, conn := range peerConn.send {
				err := conn.Encode(peerMsg)
				if err != io.EOF {
					utils.Check(err, verbosity, "Error sending (brodcast) message to peer")
				}
			}
			peerConn.lock.Unlock()
		} else {
			peerConn.send[controlMsg.PeerId].Encode(peerMsg)
		}
	}
}

func ListenForConns(
	peerConn *peerConn,
	myId,
	fileId,
	listenAddr string,
	chanPeerWire chan internal.ControlMessage,
	verbosity int,
) {
	listener, err := net.Listen("tcp", listenAddr)
	utils.Check(err, verbosity, "Error in ListenForConns")
	for {
		conn, err := listener.Accept()
		utils.Check(err, verbosity, "Error in Accepting new connection")
		utils.PrintVerbose(verbosity, utils.VERBOSE, "New connection from: ", conn.RemoteAddr().String())
		gobSend := gob.NewEncoder(conn)
		gobReceive := gob.NewDecoder(conn)
		peerId, err := PerfomHandshake(gobSend, gobReceive, myId, fileId, verbosity)
		utils.Check(err, verbosity, "Error in Handshake")
		peerConn.lock.Lock()
		peerConn.conns[peerId] = conn
		peerConn.send[peerId] = gobSend
		peerConn.receive[peerId] = gobReceive
		chanPeerWire <- internal.ControlMessage{
			Opcode:  internal.NEW_CONNECTION,
			PeerId:  peerId,
			Payload: nil,
		}
		peerConn.lock.Unlock()
		go ListenForMessages(peerConn, peerId, chanPeerWire, verbosity)
	}
}

func PerfomHandshake(
	connSend *gob.Encoder,
	connRecv *gob.Decoder,
	myId,
	fileId string,
	verbosity int,
) (string, error) {
	myHandShake := internal.HandShake{
		Pstr:   internal.PROTOCOL_ID,
		IdHash: fileId,
		PeerId: myId,
	}

	peerHandShake := internal.HandShake{}
	err := connSend.Encode(myHandShake)
	utils.Check(err, verbosity, "Error sending Handshake")
	err = connRecv.Decode(&peerHandShake)
	utils.Check(err, verbosity, "Error receiving Handshake")
	utils.PrintVerbose(verbosity, utils.DEBUG, "Handshake sucessful with peer: ", peerHandShake.PeerId[:5])
	if peerHandShake.Pstr != internal.PROTOCOL_ID || fileId != peerHandShake.IdHash {
		return "", fmt.Errorf("handshake failed")
	}
	return peerHandShake.PeerId, nil
}

func MessageOpcode(msg internal.Message) int {
	switch msg.Data.(type) {
	case internal.Have:
		return internal.HAVE
	case internal.Bitfield:
		return internal.BITFIELD
	case internal.Request:
		return internal.REQUEST
	case internal.Piece:
		return internal.PIECE
	case internal.HelloDebug:
		return internal.HELLO
	default:
		panic("Unknown message type received!")
	}
}
