package tracker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rafaelbarbeta/MicroTorr/pkg/utils"
)

type Swarm struct {
	IdHash string
	Peers  map[string]Peer
}

// Peers also should send a request with their status ("started," "stopped" or "completed")
type Peer struct {
	Ip   string
	Port int
	Id   string
}

const (
	ALIVE_TIMER = 30 * time.Second
)

var (
	Swarms        = make(map[string]Swarm)
	TimerChannels = make(map[string]*chan bool)
)

func Announce(w http.ResponseWriter, r *http.Request) {
	verbosity := utils.DEBUG // Verbosity is fixed for tracker requests for now
	fmt.Println("Announce")
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	queryParams := r.URL.Query()
	utils.PrintVerbose(3, utils.VERBOSE, "Received request: ", r.URL)
	swarmId := queryParams.Get("swarmId")
	peerId := queryParams.Get("peerId")
	ipv4 := queryParams.Get("ip")
	port, err := strconv.Atoi(queryParams.Get("port"))
	if err != nil {
		utils.PrintVerbose(verbosity, utils.CRITICAL, ipv4, ":Invalid port: Not a Number")
		http.Error(w, "Invalid port: Not a Number", http.StatusBadRequest)
		return
	}
	event := queryParams.Get("event")
	if swarmId == "" || peerId == "" || ipv4 == "" || port == 0 || event == "" {
		utils.PrintVerbose(1, utils.CRITICAL, ipv4, ":Missing required parameters")
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	swarm, exist := Swarms[swarmId]
	if !exist {
		utils.PrintVerbose(verbosity, utils.INFORMATION, ipv4, ":New Swarm created with ID: ", swarmId)
		swarm = Swarm{IdHash: swarmId, Peers: make(map[string]Peer)}
		swarm.Peers[peerId] = Peer{Ip: ipv4, Port: port, Id: peerId}
		Swarms[swarmId] = swarm
	}

	switch event {
	case "started":
		utils.PrintVerbose(verbosity, utils.VERBOSE, ipv4, ":Peer entered the swarm: ", swarmId)
		startPeerTimer(swarmId, peerId, verbosity)
		swarm.Peers[peerId] = Peer{Ip: ipv4, Port: port, Id: peerId}
		swarmJson, error := json.Marshal(swarm)
		if error != nil {
			panic("Error marshalling swarm to JSON")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(swarmJson)
	case "stopped", "completed":
		utils.PrintVerbose(verbosity, utils.VERBOSE, ipv4, ":Peer exited the swarm")
		delete(swarm.Peers, peerId)
		delete(TimerChannels, swarmId+peerId)
		w.Write([]byte("Peer exited the swarm"))
	case "alive":
		utils.PrintVerbose(verbosity, utils.DEBUG, ipv4, ":Peer is alive")
		chanPeer, ok := TimerChannels[swarmId+peerId]
		if ok {
			*chanPeer <- true
			startPeerTimer(swarmId, peerId, verbosity)
		} else {
			http.Error(w, "Peer is not in this swarm", http.StatusBadRequest)
		}
	default:
		utils.PrintVerbose(verbosity, utils.CRITICAL, ipv4, ":Sent an invalid event!")
		http.Error(w, "Invalid event", http.StatusBadRequest)
	}

	utils.PrintVerbose(verbosity, utils.DEBUG, "Swarm now: ", Swarms[swarmId])
}

func startPeerTimer(swarmId, peerId string, verbosity int) {
	peerTimer := time.NewTimer(ALIVE_TIMER)
	stopChannel := make(chan bool)
	TimerChannels[swarmId+peerId] = &stopChannel
	go func() {
		select {
		case <-peerTimer.C:
			utils.PrintVerbose(verbosity, utils.CRITICAL, peerId, ":Peer timed out")
			delete(Swarms[swarmId].Peers, peerId)
			delete(TimerChannels, swarmId+peerId)
		case <-stopChannel:
			utils.PrintVerbose(verbosity, utils.DEBUG, peerId, ":Peer timer stopped")
			break
		}
	}()
}
