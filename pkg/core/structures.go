package core

import (
	"fmt"
	"sync"

	"github.com/rafaelbarbeta/MicroTorr/pkg/utils"
)

// Internal Structures
type SyncPeerPieces struct {
	Have  map[string][]bool
	Speed map[string]float64
	Lock  sync.RWMutex
}

type PiecesBytes struct {
	Pieces [][]byte
	Have   []bool
}

type TorrentControl struct {
	DownloadingFromId string
	UploadingToIds    []string
	Lock              sync.RWMutex
}

/*
Returns the rarest pieces and the peers that have them.

	rareRank = -1 will return the rarest pieces among all peers,
	otherwise it will return the rarest pieces that are owned by
	 at least rareRank peers
	 returns List of pieces indexes, paired with their peers, and the minimum rare rank
*/
func (sp *SyncPeerPieces) RarestPieces(rareRank int) ([]int, [][]string, int) {
	sp.Lock.Lock()
	rarities := make([]int, len(sp.Have))
	peerHasPiece := make([][]string, len(sp.Have))
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
	var minRarity int
	if rareRank == -1 {
		minRarity = utils.Min(rarities)
	} else {
		minRarity = rareRank
	}

	for i := range rarities {
		if rarities[i] == minRarity {
			rarePieces = append(rarePieces, i)
			peerHasRarePiece = append(peerHasRarePiece, peerHasPiece[i])
		}
	}

	return rarePieces, peerHasRarePiece, minRarity
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
		return utils.RandomChoiceString(peers)
	} else {
		return quickestPeer
	}
}

func (sp *SyncPeerPieces) SetSpeed(peerId string, speed float64) {
	sp.Lock.Lock()
	sp.Speed[peerId] = speed
	sp.Lock.Unlock()
}

func (sp *SyncPeerPieces) AddPeer(peerId string, numberOfPieces int) {
	sp.Lock.Lock()
	sp.Have[peerId] = make([]bool, numberOfPieces)
	sp.Speed[peerId] = -1
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

func (p *PiecesBytes) GetPiece(index int) ([]byte, error) {
	if !p.Have[index] {
		return nil, fmt.Errorf("piece %d not found", index)
	}
	return p.Pieces[index], nil
}

func (p *PiecesBytes) AddPiece(piece []byte, index int) {
	p.Pieces[index] = piece
	p.Have[index] = true
}
