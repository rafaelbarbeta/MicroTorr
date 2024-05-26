package trackercontroller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

func InitTrackerController(url, id, swarmId, ip, port string, verbosity int, stopChan chan bool) {
	timer := time.NewTimer(tracker.ALIVE_TIMER - 15*time.Second)
	for {
		select {
		case <-timer.C:
			KeepAlive(url, id, swarmId, ip, port, verbosity)
			timer.Reset(tracker.ALIVE_TIMER - 15*time.Second)
		case <-stopChan:
			timer.Stop()
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
