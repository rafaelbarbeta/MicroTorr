package core

import (
	"os"
	"sync"

	"github.com/rafaelbarbeta/MicroTorr/pkg/messages"
	"github.com/rafaelbarbeta/MicroTorr/pkg/utils"
)

func HandleSignals(
	chanSignal chan os.Signal,
	chanTracker chan messages.ControlMessage,
	wait *sync.WaitGroup,
	verbosity int,
) {
	sig := <-chanSignal
	utils.PrintVerbose(verbosity, utils.CRITICAL, "Received: ", sig)
	utils.PrintVerbose(verbosity, utils.CRITICAL, "Alerting tracker and stopping execution...")
	chanTracker <- messages.ControlMessage{
		Opcode:  messages.TRACKER_STOPPED,
		PeerId:  "",
		Payload: nil,
	}
	<-chanTracker
	wait.Done()
}
