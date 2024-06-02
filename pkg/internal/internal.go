package internal

const (
	PROTOCOL_ID = "MICROTORRv1"
	// Opcodes
	NEW_CONNECTION = iota
	DEAD_CONNECTION
	HANDSHAKE
	HAVE
	BITFIELD
	REQUEST
	REJECT
	PIECE
	HELLO
	EXIT
)

// Socket Messages
type Message struct {
	Data interface{}
}

type HandShake struct {
	Pstr   string
	IdHash string
	PeerId string
}

type Have struct {
	PieceIndex int
}

type Bitfield struct {
	Bitfield []bool
}

type Request struct {
	PieceIndex int
	//Begin      int Depois...
	//Length     int
}

type Reject struct {
	PieceIndex int
}

type Piece struct {
	PieceIndex int
	//Offset     int
	Data []byte
}

type HelloDebug struct {
	Msg string
}

// Messages send between go routines
type ControlMessage struct {
	Opcode  int
	PeerId  string
	Payload interface{} //payload depends on opcode
}
