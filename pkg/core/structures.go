package core

import (
	"fmt"
	"math"
	"sync"

	"github.com/rafaelbarbeta/MicroTorr/pkg/messages"
	"github.com/rafaelbarbeta/MicroTorr/pkg/utils"
)

// messages Structures
type SyncPeerPieces struct {
	Have  map[string][]bool
	Speed map[string]float64
	Lock  sync.RWMutex
}

type PiecesBytes struct {
	Pieces [][]byte
	Hash   []string
	Have   []bool
}

type SeedMode struct {
	SeedFile string
	active   bool
	auto     bool
}

/*
Returns the rarest pieces and the peers that have them.

	rareRank = -1 will return the rarest pieces among all peers,
	otherwise it will return the rarest pieces that are owned by
	 at least rareRank peers
	 returns List of pieces indexes, paired with their peers, and the minimum rare rank
*/
func (sp *SyncPeerPieces) RarestPieces(PieceBytes *PiecesBytes, numberOfPieces int) ([]int, [][]string) {
	sp.Lock.Lock()
	rarities := make([]int, numberOfPieces)
	peerHasPiece := make([][]string, numberOfPieces)
	rarePieces := make([]int, 0)
	peerHasRarePiece := make([][]string, 0)
	for peer, have := range sp.Have {
		for i := range have {
			if have[i] {
				rarities[i]++
				peerHasPiece[i] = append(peerHasPiece[i], peer)
			}
		}
	}
	sp.Lock.Unlock()
	minRarity := utils.Min(rarities)

	for i := range rarities {
		if rarities[i] == minRarity && !PieceBytes.Have[i] {
			rarePieces = append(rarePieces, i)
			peerHasRarePiece = append(peerHasRarePiece, peerHasPiece[i])
		}
	}

	return rarePieces, peerHasRarePiece
}

func (sp *SyncPeerPieces) QuickestPeer(peers []string) string {
	maxSpeed := float64(-1)
	quickestPeer := ""

	sp.Lock.Lock()
	for _, peerId := range peers {
		if sp.Speed[peerId] > maxSpeed {
			maxSpeed = sp.Speed[peerId]
			quickestPeer = peerId
		}
	}
	sp.Lock.Unlock()
	if quickestPeer == "" {
		// Randomly choose a peer in peers
		chosenPeer, _ := utils.RandomChoiceString(peers)
		return chosenPeer
	} else {
		return quickestPeer
	}
}

func (sp *SyncPeerPieces) HasSeeder() bool {
	for _, have := range sp.Have {
		hasAllPieces := true
		for _, truthValue := range have {
			if !truthValue {
				hasAllPieces = false
				break
			}
		}
		if hasAllPieces {
			return true
		}
	}
	return false
}

func (sp *SyncPeerPieces) SetSpeed(peerId string, speed float64) {
	sp.Lock.Lock()
	sp.Speed[peerId] = speed
	sp.Lock.Unlock()
}

func (sp *SyncPeerPieces) AddPeer(peerId string, numberOfPieces int) {
	sp.Lock.Lock()
	sp.Have[peerId] = make([]bool, numberOfPieces)
	sp.Speed[peerId] = math.MaxInt64 //
	sp.Lock.Unlock()
}

func (sp *SyncPeerPieces) DeletePeer(peerId string) {
	sp.Lock.Lock()
	delete(sp.Have, peerId)
	delete(sp.Speed, peerId)
	sp.Lock.Unlock()
}

func (sp *SyncPeerPieces) AddPiece(peerId string, index int) {
	sp.Lock.Lock()
	sp.Have[peerId][index] = true
	sp.Lock.Unlock()
}

func (sp *SyncPeerPieces) SetBitfield(peerId string, bitfield messages.Bitfield) {
	sp.Lock.Lock()
	sp.Have[peerId] = bitfield.Bitfield
	sp.Lock.Unlock()
}

func (p *PiecesBytes) GetPiece(index int) ([]byte, error) {
	if !p.Have[index] {
		return nil, fmt.Errorf("piece %d not found", index)
	}
	return p.Pieces[index], nil
}

func (p *PiecesBytes) AddHash(sha1Hash string, index int) {
	p.Hash[index] = sha1Hash
}

func (p *PiecesBytes) AddPiece(piece []byte, index int) {
	p.Pieces[index] = piece
	p.Have[index] = true
}
