// Copyright ©2020 BlinnikovAA. All rights reserved.
// This file is part of yagogame.
//
// yagogame is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// yagogame is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with yagogame.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/yagoggame/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GameMode int

const (
	noGame GameMode = iota
	waitJoin
	waitTurn
	performTurn
	gameOver
)

// gameState is type to hold current state of the game.
type gameState struct {
	client api.GoGameClient
	//currentMode
	currentMode GameMode
	//chanel to await for server continious actions
	gameWaiter <-chan interface{}
	//cancel function for server continious actions
	cancel context.CancelFunc
	//greetings message
	msg string
}

func (state *gameState) processUserCommands(cmdLines <-chan string, quit <-chan interface{}) {
	process := true
	for process == true {
		select {
		//parse user commands.
		case txt := <-cmdLines:
			process = state.processKey(txt)
		//wait for continious actions.
		case rez := <-state.gameWaiter:
			state.releaseWaitingResources()
			state.processWaitResult(rez)
		//OS quit signal interseptor.
		case <-quit:
			process = false
		}
		// print a messagi according to current state.
		state.printInvitation()
	}
}

// printInvitation prints invitation before accept user input.
func (state *gameState) printInvitation() {
	msg := "IF YOU CAN SEE IT: The author is an IDIOT"
	switch state.currentMode {
	case noGame:
		msg = fmt.Sprintln("\nSelect a type of game:\n [q]: - quit from the Lobby.\n [j]: - Game with someone on the standart field.")
	case waitJoin:
		msg = fmt.Sprintln("\nWaiting for the game to start:\n [q]: - quit from the Lobby.")
	case waitTurn:
		msg = fmt.Sprintln("\nWaiting for the turn:\n [q]: - quit from the Lobby.\n [e]: - exith this Game.")
	case performTurn:
		msg = fmt.Sprintln("\nPlease, make a turn:\n [q]: - quit from the Lobby.\n [e]: - exith this Game.\n [xxx yyy]: - enter coordinates to make a turn.")
	case gameOver:
		msg = fmt.Sprintln("\nThe Game is over:\n [q]: - quit from the Lobby.\n [e]: - exith this Game.")
	}
	fmt.Println(msg)
}

//releaseWaitingResources releases resources, if any.
func (state *gameState) releaseWaitingResources() {
	state.gameWaiter = nil
	if state.cancel != nil {
		state.cancel()
		state.cancel = nil
	}
}

//releaseGameResources releases game specific resources.
func (state *gameState) releaseGameResources() {
	if state.currentMode == waitTurn || state.currentMode == performTurn || state.currentMode == gameOver {
		if _, err := state.client.LeaveTheGame(context.Background(), &api.EmptyMessage{}); err != nil {
			st := status.Convert(err)
			fmt.Println("Error, while leaving a game: %v: %s", st.Code(), st.Message())
		} else {
			state.currentMode = noGame
		}
	}
	fmt.Println("Leave The game...")
}

// processWaitResult waits of waiting function result and process it.
func (state *gameState) processWaitResult(rez interface{}) {
	switch state.currentMode {
	case waitJoin:
		if err, ok := rez.(error); ok {
			state.currentMode = noGame
			fmt.Println(err)
			break
		}
		state.currentMode = waitTurn
		state.gameWaiter, state.cancel = waitTurnBegin(state.client)
	case waitTurn:
		if err, ok := rez.(error); ok {
			state.currentMode = gameOver
			fmt.Println(err)
			break
		}
		state.currentMode = performTurn
	}

}

func (state *gameState) checkTurnData(txt string) (x, y, n int) {
	if state.currentMode == performTurn && len(txt) > 1 {
		if ln, err := fmt.Sscanf(txt, "%d %d", &x, &y); err == nil {
			n = ln
		}
	}
	return x, y, n
}

// processKey processes scanned line to find command allowed in current mode.
func (state *gameState) processKey(txt string) bool {
	x, y, n := state.checkTurnData(txt)

	switch {
	case txt == "q":
		return false

	case txt == "j" && state.currentMode == noGame:
		state.gameWaiter, state.cancel = waitJoinGame(state.client)
		state.currentMode = waitJoin

	case txt == "e" && (state.currentMode == waitTurn || state.currentMode == performTurn || state.currentMode == gameOver):
		state.releaseGameResources()

	case txt == "e" && (state.currentMode == waitTurn || state.currentMode == performTurn || state.currentMode == gameOver):
		state.releaseGameResources()

	case state.currentMode == performTurn && n == 2:
		_, err := state.client.MakeTurn(context.Background(), &api.TurnMessage{X: int64(x), Y: int64(y)})
		if err != nil {
			st := status.Convert(err)
			if st.Code() == codes.InvalidArgument {
				fmt.Println(st.Message())
				break
			}
			fmt.Println("Error, while leaving a game: %v: %s", st.Code(), st.Message())
		}
		state.currentMode = waitTurn
		state.gameWaiter, state.cancel = waitTurnBegin(state.client)

	default:
		fmt.Printf("no command %q in current mode \n", txt)
	}
	return true
}

// manageGame selects a game type, initiate and manage it.
func manageGame(client api.GoGameClient, quit <-chan interface{}) error {
	fmt.Println("Whelcome to a Go game")
	state := &gameState{currentMode: noGame, client: client}
	defer state.releaseWaitingResources()
	defer state.releaseGameResources()
	state.printInvitation()

	//asynchronous scanning of user commands.
	stopScan := make(chan interface{})
	cmdLines := scanner(stopScan)
	defer func(stopScan chan<- interface{}) {
		close(stopScan)
	}(stopScan)

	state.processUserCommands(cmdLines, quit)

	return nil
}

// scanner scans input into the chanel in separate goroutine
func scanner(stopScan <-chan interface{}) <-chan string {
	lines := make(chan string)
	go func(stopScan <-chan interface{}, lines chan<- string) {
		scanner := bufio.NewScanner(os.Stdin)

	ENDGAME:
		for scanner.Scan() {
			select {
			// scanning is already not needed
			case <-stopScan:
				break ENDGAME
			//continue ccanning
			default:
				break
			}
			//get scan result
			txt := scanner.Text()
			lines <- txt
		}
	}(stopScan, lines)
	return lines
}

// waitJoinGame initiates joining to a game.
// returns chanel to report on success or failure and
// function of cancellation.
func waitJoinGame(client api.GoGameClient) (<-chan interface{}, context.CancelFunc) {
	waitEnded := make(chan interface{})
	ctx, cancel := context.WithCancel(context.Background())

	go func(chan<- interface{}) {
		_, err := client.JoinTheGame(ctx, &api.EmptyMessage{})
		if err != nil {
			st := status.Convert(err)
			waitEnded <- fmt.Errorf("can't join a game: %v: %s", st.Code(), st.Message())
		}
		close(waitEnded)
	}(waitEnded)

	return waitEnded, cancel
}

// waitTurnBegin initiates awaiting of player's turn.
// returns chanel to report on success or failure and
// function of cancellation.
func waitTurnBegin(client api.GoGameClient) (<-chan interface{}, context.CancelFunc) {
	waitEnded := make(chan interface{})
	ctx, cancel := context.WithCancel(context.Background())

	go func(chan<- interface{}) {
		_, err := client.WaitTheTurn(ctx, &api.EmptyMessage{})
		if err != nil {
			st := status.Convert(err)
			waitEnded <- fmt.Errorf("can't wait a turn: %v: %s", st.Code(), st.Message())
		}
		close(waitEnded)
	}(waitEnded)

	return waitEnded, cancel
}
