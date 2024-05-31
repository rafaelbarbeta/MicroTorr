package core

import (
	"fmt"

	"github.com/rafaelbarbeta/MicroTorr/pkg/internal"
)

func HandleHave(msg internal.ControlMessage) {
	fmt.Println("Have handler not implemented")
}

func HandleBitfield(msg internal.ControlMessage) {
	fmt.Println("Bitfield handler not implemented")
}

func HandleRequest(msg internal.ControlMessage) {
	fmt.Println("Request handler not implemented")
}

func HandleReject(msg internal.ControlMessage) {
	fmt.Println("Reject handler not implemented")
}

func HandlePiece(msg internal.ControlMessage) {
	fmt.Println("Piece handler not implemented")
}
