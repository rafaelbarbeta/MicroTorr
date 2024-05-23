package tracker

import (
	"encoding/json"
	"net/http"
	"strconv"
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

var (
	Swarms = make(map[string]Swarm)
)

func Announce(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	queryParams := r.URL.Query()
	swarmId := queryParams.Get("swarmId")
	peer_id := queryParams.Get("peerId")
	ipv4 := queryParams.Get("ip")
	port, err := strconv.Atoi(queryParams.Get("port"))
	if err != nil {
		http.Error(w, "Invalid port: Not a Number", http.StatusBadRequest)
		return
	}
	event := queryParams.Get("event")
	if swarmId == "" || peer_id == "" || ipv4 == "" || port == 0 || event == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	swarm, exist := Swarms[swarmId]
	if !exist {
		swarm = Swarm{IdHash: swarmId, Peers: make(map[string]Peer)}
		swarm.Peers[peer_id] = Peer{Ip: ipv4, Port: port, Id: peer_id}
		Swarms[swarmId] = swarm
	}

	if event == "stopped" || event == "completed" {
		delete(swarm.Peers, peer_id)
		w.Write([]byte("Peer exited the swarm"))
		return
	} else if event == "started" {
		swarm.Peers[peer_id] = Peer{Ip: ipv4, Port: port, Id: peer_id}
		swarmJson, error := json.Marshal(swarm.Peers)
		if error != nil {
			panic("Error marshalling swarm to JSON")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(swarmJson)
		return
	} else {
		http.Error(w, "Invalid event", http.StatusBadRequest)
		return
	}
}
