package internal

import "sync"

// Internal Structures
type SyncPeerPieces struct {
	Piece [][]string // PieceId as row index to PeerId who have this piece
	Lock  sync.RWMutex
}

type SyncPeerSpeeds struct {
	Speed map[string]float64
	Lock  sync.RWMutex
}

type SyncPieceRarity struct {
	Rarity [][]int // row index is the count of peers who have this piece, column index is the piece index
	Lock   sync.RWMutex
}

// Socket Messages
type HandShake struct {
	Pstrlen  byte
	Pstr     string
	Reserved [8]byte
	IdHash   string
	PeerId   string
}

type Have struct {
	Id         byte
	PieceIndex int
}

type Bitfield struct {
	Id       byte
	Bitfield []byte
}

type Request struct {
	Id         byte
	PieceIndex int
	Begin      int
	Length     int
}

type Reject struct {
	Id         byte
	PieceIndex int
}

type PartialPiece struct {
	Id         byte
	PieceIndex int
	Offset     int
	Data       []byte
}

// Messages send between go routines
type ControlMessage struct {
	Opcode  int
	payload interface{} //payload depends on opcode
}
