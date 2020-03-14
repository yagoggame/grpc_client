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

package client

import (
	"fmt"
	"strings"

	"github.com/yagoggame/api"
)

type positionState int

const (
	empty positionState = iota
	blackChip
	whiteChip
)

var symbols = map[positionState]string{
	empty:     "┼",
	blackChip: "●",
	whiteChip: "○",
}

func stringFromGameData(state *api.State) string {
	size := int(state.GetSize())
	desk := make([][]positionState, size)
	for i := range desk {
		desk[i] = make([]positionState, size)
	}
	fillMarkers(desk, state)

	line := ""
	for y := range desk {
		line += strings.Repeat(" │ ", size) + "\n"
		subline := ""

		for _, ps := range desk[y] {
			subline += "─" + symbols[ps] + "─"
		}
		line += subline + "\n"
	}

	line += strings.Repeat(" │ ", size) + "\n"
	line += fmt.Sprintf("komi is %f\n", state.GetKomi())
	line += fmt.Sprintf("Black situation: %s\n", describeGamerSituation(state.GetBlack()))
	line += fmt.Sprintf("White situation: %s\n", describeGamerSituation(state.GetWhite()))
	return line
}

func describeGamerSituation(colourSituation *api.State_ColourState) string {
	rez := fmt.Sprintf("Chips in cup: %3d ", colourSituation.GetChipsInCap())
	rez += fmt.Sprintf("Chips cuptured: %3d ", colourSituation.GetChipsCaptured())
	rez += fmt.Sprintf("Chips cuptured: %5.1f", colourSituation.GetScores())
	return rez
}

func fillMarkers(desk [][]positionState, state *api.State) {
	for _, p := range state.GetBlack().GetChipsOnBoard() {
		desk[p.Y-1][p.X-1] = blackChip
	}
	for _, p := range state.GetWhite().GetChipsOnBoard() {
		desk[p.Y-1][p.X-1] = whiteChip
	}
}
