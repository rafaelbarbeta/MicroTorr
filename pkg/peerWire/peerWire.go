package peerWire

import (
	"encoding/gob"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/conduitio/bwlimit"
	"github.com/rafaelbarbeta/MicroTorr/pkg/messages"
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
	chanPeerWire, chanCore chan messages.ControlMessage,
	wait *sync.WaitGroup,
	maxDownSpeed, maxUpSpeed, verbosity int,
) {
	gob.Register(messages.HandShake{})
	gob.Register(messages.Have{})
	gob.Register(messages.Bitfield{})
	gob.Register(messages.Request{})
	gob.Register(messages.Piece{})
	gob.Register(messages.HelloDebug{})

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
		dialer := bwlimit.NewDialer(&net.Dialer{}, bwlimit.Byte(maxUpSpeed)*bwlimit.KB, bwlimit.Byte(maxDownSpeed)*bwlimit.KB)
		conn, err := dialer.Dial("tcp", peer.Ip+":"+strconv.Itoa(peer.Port))
		utils.Check(err, verbosity, "Error connecting to peer: ", peer.Id[:5])
		peerConn.conns[peer.Id] = conn
		peerConn.send[peer.Id] = gob.NewEncoder(conn)
		peerConn.receive[peer.Id] = gob.NewDecoder(conn)
		peerId, err := PerfomHandshake(peerConn.send[peer.Id],
			peerConn.receive[peer.Id],
			myId,
			swarm.IdHash,
			verbosity)
		utils.Check(err, verbosity, "Error performing Handshake with peer: ", peer.Id[:5])
		utils.PrintVerbose(verbosity, utils.VERBOSE, "Handshake with peer: ", peer.Id[:5], "sucessful")
		chanPeerWire <- messages.ControlMessage{
			Opcode:  messages.NEW_CONNECTION,
			PeerId:  peerId,
			Payload: nil,
		}
		go ListenForMessages(&peerConn, peer.Id, chanPeerWire, verbosity)
	}

	// Start listener for new connections
	go ListenForConns(
		&peerConn,
		myId,
		swarm.IdHash,
		swarm.Peers[myId].Ip+":"+strconv.Itoa(swarm.Peers[myId].Port),
		chanPeerWire,
		maxDownSpeed,
		maxUpSpeed,
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
	chanPeerWire chan messages.ControlMessage,
	verbosity int,
) {
	utils.PrintVerbose(verbosity, utils.DEBUG, "Starting listening for messages from peer: ", peerId[:5])
	var msg messages.Message
	for {
		err := peerConn.receive[peerId].Decode(&msg)
		// verificar desconexão ou erro de envio
		if err != nil {
			peerConn.lock.Lock()
			DisconnectPeer(peerConn, chanPeerWire, peerId, verbosity)
			peerConn.lock.Unlock()
			return
		}
		opcode := MessageOpcode(msg)
		utils.PrintVerbose(verbosity, utils.DEBUG, "Message from peer: ", peerId[:5], " opcode: ", opcode)
		chanPeerWire <- messages.ControlMessage{
			Opcode:  opcode,
			PeerId:  peerId,
			Payload: msg.Data,
		}
	}
}

func ListenForCoreMessages(
	peerConn *peerConn,
	chanCore chan messages.ControlMessage,
	wait *sync.WaitGroup,
	verbosity int,
) {
	var controlMsg messages.ControlMessage
	var peerMsg messages.Message
	for {
		controlMsg = <-chanCore
		peerMsg = messages.Message{Data: controlMsg.Payload}
		if controlMsg.PeerId == "" { // Empty string is used to broadcast message
			peerConn.lock.Lock()
			for _, conn := range peerConn.send {
				err := conn.Encode(peerMsg)
				if err != nil {
					DisconnectPeer(peerConn, chanCore, controlMsg.PeerId, verbosity)
				}
			}
			peerConn.lock.Unlock()
		} else {
			err := peerConn.send[controlMsg.PeerId].Encode(peerMsg)
			if err != nil {
				peerConn.lock.Lock()
				DisconnectPeer(peerConn, chanCore, controlMsg.PeerId, verbosity)
				peerConn.lock.Unlock()
			}
		}
	}
}

func ListenForConns(
	peerConn *peerConn,
	myId, fileId, listenAddr string,
	chanPeerWire chan messages.ControlMessage,
	maxDownSpeed, maxUpSpeed, verbosity int,
) {
	listener, err := net.Listen("tcp", listenAddr)
	utils.Check(err, verbosity, "Error in ListenForConns")
	listenerLimited := bwlimit.NewListener(listener, bwlimit.Byte(maxUpSpeed)*bwlimit.KB, bwlimit.Byte(maxUpSpeed)*bwlimit.KB)
	for {
		conn, err := listenerLimited.Accept()
		utils.Check(err, verbosity, "Error in Accepting new connection")
		utils.PrintVerbose(verbosity, utils.VERBOSE, "New connection from: ", conn.RemoteAddr().String())
		gobSend := gob.NewEncoder(conn)
		gobReceive := gob.NewDecoder(conn)
		peerId, err := PerfomHandshake(gobSend, gobReceive, myId, fileId, verbosity)
		if err != nil {
			utils.PrintVerbose(verbosity, utils.CRITICAL, "Error in Perfoming Handshake: ", err)
			continue
		}
		peerConn.lock.Lock()
		peerConn.conns[peerId] = conn
		peerConn.send[peerId] = gobSend
		peerConn.receive[peerId] = gobReceive
		chanPeerWire <- messages.ControlMessage{
			Opcode:  messages.NEW_CONNECTION,
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
	myId, fileId string,
	verbosity int,
) (string, error) {
	myHandShake := messages.HandShake{
		Pstr:   messages.PROTOCOL_ID,
		IdHash: fileId,
		PeerId: myId,
	}

	peerHandShake := messages.HandShake{}
	err := connSend.Encode(myHandShake)
	if err != nil {
		return "", fmt.Errorf("error sending handshake")
	}
	err = connRecv.Decode(&peerHandShake)
	if err != nil {
		return "", fmt.Errorf("error receiving handshake")
	}
	if peerHandShake.Pstr != messages.PROTOCOL_ID || fileId != peerHandShake.IdHash {
		return "", fmt.Errorf("handshake failed: protocol id or file id mismatch")
	}
	utils.PrintVerbose(verbosity, utils.DEBUG, "Handshake sucessful with peer: ", peerHandShake.PeerId[:5])
	return peerHandShake.PeerId, nil
}

func MessageOpcode(msg messages.Message) int {
	switch msg.Data.(type) {
	case messages.Have:
		return messages.HAVE
	case messages.Bitfield:
		return messages.BITFIELD
	case messages.Request:
		return messages.REQUEST
	case messages.Piece:
		return messages.PIECE
	case messages.HelloDebug:
		return messages.HELLO
	default:
		panic("Unknown message type received!")
	}
}

func DisconnectPeer(
	peerConn *peerConn,
	chanPeerWire chan messages.ControlMessage,
	peerId string,
	verbosity int,
) {
	utils.PrintVerbose(verbosity, utils.CRITICAL, "Peer: ", peerId[:5], "disconnected!")
	delete(peerConn.conns, peerId)
	delete(peerConn.send, peerId)
	delete(peerConn.receive, peerId)
	chanPeerWire <- messages.ControlMessage{
		Opcode:  messages.DEAD_CONNECTION,
		PeerId:  peerId,
		Payload: nil,
	}
}
