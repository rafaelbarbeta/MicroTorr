package tracker

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
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
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	queryParams := r.URL.Query()
	swarmId := queryParams.Get("swarmId")
	peerId := queryParams.Get("peerId")
	ipv4 := queryParams.Get("ip")
	port, err := strconv.Atoi(queryParams.Get("port"))
	if err != nil {
		http.Error(w, "Invalid port: Not a Number", http.StatusBadRequest)
		return
	}
	event := queryParams.Get("event")
	if swarmId == "" || peerId == "" || ipv4 == "" || port == 0 || event == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	swarm, exist := Swarms[swarmId]
	if !exist {
		swarm = Swarm{IdHash: swarmId, Peers: make(map[string]Peer)}
		swarm.Peers[peerId] = Peer{Ip: ipv4, Port: port, Id: peerId}
		Swarms[swarmId] = swarm
	}

	switch event {
	case "started":
		startPeerTimer(swarmId, peerId)
		swarm.Peers[peerId] = Peer{Ip: ipv4, Port: port, Id: peerId}
		swarmJson, error := json.Marshal(swarm)
		if error != nil {
			panic("Error marshalling swarm to JSON")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(swarmJson)
	case "stopped", "completed":
		delete(swarm.Peers, peerId)
		delete(TimerChannels, swarmId+peerId)
		w.Write([]byte("Peer exited the swarm"))
	case "alive":
		chanPeer, ok := TimerChannels[swarmId+peerId]
		if ok {
			*chanPeer <- true
			startPeerTimer(swarmId, peerId)
		} else {
			http.Error(w, "Peer is not in this swarm", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Invalid event", http.StatusBadRequest)
	}
}

func startPeerTimer(swarmId, peerId string) {
	peerTimer := time.NewTimer(ALIVE_TIMER)
	stopChannel := make(chan bool)
	TimerChannels[swarmId+peerId] = &stopChannel
	go func() {
		select {
		case <-peerTimer.C:
			delete(Swarms[swarmId].Peers, peerId)
			delete(TimerChannels, swarmId+peerId)
		case <-stopChannel:
			break
		}
	}()
}
