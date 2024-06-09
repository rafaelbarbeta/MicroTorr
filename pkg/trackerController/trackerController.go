package trackercontroller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rafaelbarbeta/MicroTorr/pkg/messages"
	"github.com/rafaelbarbeta/MicroTorr/pkg/tracker"
	"github.com/rafaelbarbeta/MicroTorr/pkg/utils"
)

func GetTrackerInfo(url, id, swarmId, ip, port string, verbosity int) tracker.Swarm {
	urlParameters := url + fmt.Sprintf(
		"/announce?peerId=%s&swarmId=%s&ip=%s&port=%s&event=started",
		id, swarmId, ip, port)

	utils.PrintVerbose(verbosity, utils.VERBOSE, "Requesting: ", urlParameters)
	response, err := http.Get(urlParameters)
	utils.Check(err, verbosity, "Error requesting ", urlParameters)
	var swarm tracker.Swarm
	err = json.NewDecoder(response.Body).Decode(&swarm)
	utils.Check(err, verbosity, "Error decoding JSON response")

	return swarm
}

func InitTrackerController(url, id, swarmId, ip, port string, verbosity int, chanTracker chan messages.ControlMessage) {
	timer := time.NewTimer(tracker.ALIVE_TIMER - 15*time.Second)
	for {
		select {
		case <-timer.C:
			KeepAlive(url, id, swarmId, ip, port, verbosity)
			timer.Reset(tracker.ALIVE_TIMER - 15*time.Second)
		case msg := <-chanTracker:
			switch msg.Opcode {
			case messages.TRACKER_COMPLETED:
				DownloadCompleted(url, id, swarmId, ip, port, verbosity)
				chanTracker <- messages.ControlMessage{
					Opcode:  messages.EXIT,
					PeerId:  "",
					Payload: nil,
				}
			case messages.TRACKER_STOPPED:
				DownloadStopped(url, id, swarmId, ip, port, verbosity)
				chanTracker <- messages.ControlMessage{
					Opcode:  messages.EXIT,
					PeerId:  "",
					Payload: nil,
				}
			}
			return
		}
	}
}

func KeepAlive(url, id, swarmId, ip, port string, verbosity int) {
	urlParameters := url + fmt.Sprintf(
		"/announce?peerId=%s&swarmId=%s&ip=%s&port=%s&event=alive",
		id, swarmId, ip, port)

	utils.PrintVerbose(verbosity, utils.DEBUG, "Keeping Alive: ", urlParameters)
	_, err := http.Get(urlParameters)
	utils.Check(err, verbosity, "Error: keep alive failed!")
}

func DownloadCompleted(url, id, swarmId, ip, port string, verbosity int) {
	urlParameters := url + fmt.Sprintf(
		"/announce?peerId=%s&swarmId=%s&ip=%s&port=%s&event=completed",
		id, swarmId, ip, port)

	_, err := http.Get(urlParameters)
	utils.Check(err, verbosity, "Error: download completed failed!")
}

func DownloadStopped(url, id, swarmId, ip, port string, verbosity int) {
	urlParameters := url + fmt.Sprintf(
		"/announce?peerId=%s&swarmId=%s&ip=%s&port=%s&event=stopped",
		id, swarmId, ip, port)

	_, err := http.Get(urlParameters)
	utils.Check(err, verbosity, "Error: download stopped failed!")
}
