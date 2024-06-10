package core

import (
	"crypto/sha1"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rafaelbarbeta/MicroTorr/pkg/messages"
	"github.com/rafaelbarbeta/MicroTorr/pkg/mtorr"
	"github.com/rafaelbarbeta/MicroTorr/pkg/utils"
	"github.com/schollz/progressbar/v3"
)

func PieceRequester(
	PeerPieces *SyncPeerPieces,
	PiecesBytes *PiecesBytes,
	SeedMode *SeedMode,
	mtorrrent mtorr.Mtorrent,
	numberOfPieces int,
	chanPieceRequester, chanCore, chanTracker chan messages.ControlMessage,
	wait *sync.WaitGroup,
	waitSeeders, waitLeechers, verbosity int,
	bar *progressbar.ProgressBar,
) {
	var piecesIdx []int
	var peers [][]string
	var selectedPeer string
	var selectedPiece, selectedPieceIdx int
	var msg messages.ControlMessage
	var timeStart time.Time
	stats := DownloadStats{
		PiecesDownloaded: make([]int, 0),
		PiecesSpeed:      make([]float64, 0),
		FromPeers:        make([]string, 0),
		FromSeeder:       make([]bool, 0),
	}

	// Wait until the minimum number of seeders/leechers are in the swarm
	for PeerPieces.NumSeeders() < waitSeeders || PeerPieces.NumLeechers()+1 < waitLeechers {
		utils.PrintVerbose(verbosity, utils.DEBUG, "Seeders: ", PeerPieces.NumSeeders(), " Leechers: ", PeerPieces.NumLeechers()+1)
		utils.PrintVerbose(verbosity, utils.DEBUG, "Required Seeders: ", waitSeeders, " Required Leechers: ", waitLeechers)
		time.Sleep(WAIT_DEFAULT_TIME)
	}

	utils.PrintVerbose(verbosity, utils.VERBOSE, "Downloading pieces...")

	for {
		piecesIdx, peers = PeerPieces.RarestPieces(PiecesBytes, numberOfPieces)
		if len(piecesIdx) == 0 { // Only happens if all pieces have been downloaded already
			AssemblePieces(mtorrrent, PiecesBytes, SeedMode, chanTracker, &stats, wait, verbosity, bar)
			break
		}
		selectedPiece, selectedPieceIdx = utils.RandomChoiceInt(piecesIdx)
		// Minimum chance of chosing a random peer regardless of it being the quickest
		if utils.RandomPercentChance(OPPORTUNISTIC_CHOICE) {
			selectedPeer = PeerPieces.QuickestPeer(peers[selectedPieceIdx])
		} else {
			utils.PrintVerbose(verbosity, utils.DEBUG, "Trying a random peer instead a quick peer...")
			selectedPeer, _ = utils.RandomChoiceString(peers[selectedPieceIdx])
		}
		utils.PrintVerbose(verbosity, utils.DEBUG, "Requesting piece ", selectedPiece, " from peer ", selectedPeer)

		timeStart = time.Now() // To calculate download speed
		chanCore <- messages.ControlMessage{
			Opcode: messages.REQUEST,
			PeerId: selectedPeer,
			Payload: messages.Request{
				PieceIndex: selectedPiece,
			},
		}

		// Wait for response. Make sure it's from the requested peer
		for {
			msg = <-chanPieceRequester
			if msg.PeerId != selectedPeer && msg.Opcode == messages.DEAD_CONNECTION {
				continue
			} else if msg.PeerId != selectedPeer {
				utils.PrintVerbose(verbosity, utils.CRITICAL,
					"Received unsolicited message from ",
					selectedPeer[:5])
				continue
			} else {
				break
			}
		}
		switch msg.Opcode {
		case messages.PIECE:
			duration := time.Since(timeStart).Seconds()
			speed := float64(mtorrrent.Info.Piece_length) / duration
			go func(selectedPiece int) { //Making sure the hashes match. Not ideal, but errors out if they don't
				hashPiece := fmt.Sprintf("%x",
					sha1.Sum(msg.Payload.(messages.Piece).Data),
				)
				if hashPiece != PiecesBytes.Hash[selectedPiece] {
					utils.PrintVerbose(verbosity, utils.CRITICAL, "Piece Hash does not match!")
					wait.Done()
					os.Exit(1)
				}
			}(selectedPiece)
			PeerPieces.SetSpeed(selectedPeer, speed)
			PiecesBytes.AddPiece(msg.Payload.(messages.Piece).Data, msg.Payload.(messages.Piece).PieceIndex)
			utils.PrintVerbose(verbosity, utils.DEBUG,
				"Piece: ",
				msg.Payload.(messages.Piece).PieceIndex,
				" from: ", selectedPeer[:5],
				"speed: ",
				fmt.Sprintf("%.3f MB/s", speed/1000000.0),
			)
			stats.Update(msg.Payload.(messages.Piece).PieceIndex,
				speed, selectedPeer, PeerPieces.IsSeeder(selectedPeer))
			if verbosity != utils.DEBUG {
				bar.Add(mtorrrent.Info.Piece_length)
			}
		case messages.DEAD_CONNECTION:
			utils.PrintVerbose(verbosity, utils.DEBUG,
				"Peer", selectedPeer[:5], "cannot send piece ",
				selectedPiece,
				" because it is dead")
		default:
			panic("Unknown message type received at PieceRequester!")
		}
		chanCore <- messages.ControlMessage{
			Opcode: messages.HAVE,
			PeerId: "",
			Payload: messages.Have{
				PieceIndex: selectedPiece,
			},
		}
	}
}

func AssemblePieces(
	mtorrent mtorr.Mtorrent,
	PiecesBytes *PiecesBytes,
	SeedMode *SeedMode,
	chanTracker chan messages.ControlMessage,
	stats *DownloadStats,
	wait *sync.WaitGroup,
	verbosity int,
	bar *progressbar.ProgressBar,
) {
	if verbosity != utils.DEBUG {
		bar.Exit()
	}
	utils.PrintVerbose(verbosity, utils.VERBOSE, "All pieces downloaded. Assembling...")

	var data []byte
	for i := 0; i < len(PiecesBytes.Pieces); i++ {
		data = append(data, PiecesBytes.Pieces[i]...)
	}
	// Checks the Sha1sum
	if fmt.Sprintf("%x", sha1.Sum(data)) != mtorrent.Info.Id {
		utils.PrintVerbose(verbosity, utils.CRITICAL, "Assembled pieces SHA1 does not match with Mtorrent SHA1!.")
		utils.PrintVerbose(verbosity, utils.DEBUG, "Assembled pieces SHA1: ", fmt.Sprintf("%x", sha1.Sum(data)), " \nMtorrent SHA1:", mtorrent.Info.Id)
	} else {
		utils.PrintVerbose(verbosity, utils.INFORMATION, "Assembled pieces SHA1 matches with Mtorrent SHA1")
	}

	utils.PrintVerbose(verbosity, utils.INFORMATION, "Dumping Data...")
	err := os.WriteFile(mtorrent.Info.Name, data, 0644)
	utils.Check(err, verbosity, "Failed to write assembled data to disk")
	utils.PrintVerbose(verbosity, utils.VERBOSE, "Data dumped to disk")
	utils.PrintVerbose(verbosity, utils.CRITICAL, stats)
	if SeedMode.auto {
		utils.PrintVerbose(verbosity, utils.VERBOSE, "Changed to seeding mode")
		SeedMode.active = true
		SeedMode.SeedFile = mtorrent.Info.Name
	} else {
		utils.PrintVerbose(verbosity, utils.VERBOSE, "Exiting swarm...")
		chanTracker <- messages.ControlMessage{
			Opcode:  messages.TRACKER_COMPLETED,
			PeerId:  "",
			Payload: nil,
		}
		<-chanTracker
		wait.Done()
	}
}

func PieceUploader(
	PiecesBytes *PiecesBytes,
	mtorrent mtorr.Mtorrent,
	chanPieceUploader, chanCore chan messages.ControlMessage,
	verbosity int,
) {
	for {
		msg := <-chanPieceUploader
		go func(msg messages.ControlMessage) {
			chanCore <- messages.ControlMessage{
				Opcode: messages.PIECE,
				PeerId: msg.PeerId,
				Payload: messages.Piece{
					PieceIndex: msg.Payload.(messages.Request).PieceIndex,
					Data:       PiecesBytes.Pieces[msg.Payload.(messages.Request).PieceIndex],
				},
			}
			utils.PrintVerbose(
				verbosity, utils.DEBUG,
				"Sent piece ", msg.Payload.(messages.Request).PieceIndex,
				" to: ", msg.PeerId[:5],
			)
		}(msg)
	}
}
