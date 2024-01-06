// MIT License
//
// Copyright (c) 2023 Jakob Görgen
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"log"
	"math"
	"math/rand"
	"os"
	"time"
)

func ErrExit(err error) {
	if err == nil {
		return
	}
	fmt.Printf("%s", err)
	os.Exit(1)
}

type Velocity struct {
	X int
	Y int
}

func (me *Velocity) Equals(other Velocity) bool {
	return me.X == other.X && me.Y == other.Y
}

var NorthDir = Velocity{X: 0, Y: -1}
var SouthDir = Velocity{X: 0, Y: 1}
var EastDir = Velocity{X: 1, Y: 0}
var WestDir = Velocity{X: -1, Y: 0}

var blackWhiteStyle = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
var backStyle = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
var snailBodySytle = tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorWhite)
var wallStyle = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorBlue)
var snailHeadSytle = tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorGreen)
var foodStyle = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorRed)

type Pos struct {
	X int
	Y int
}

type Scorer struct {
	Score             int
	movesSinceLastInc int
	weight            int
	gridWidth         int
	gridHeight        int
	maxPoints         int
	OldHeadPos        Pos
	OldFoodPos        Pos
}

func (scorer *Scorer) Step() {
	scorer.movesSinceLastInc += 1
}

func (scorer *Scorer) ResetSteps() {
	scorer.movesSinceLastInc = 0
}

func (scorer *Scorer) CalculateScore() error {
	defer scorer.ResetSteps()
	if scorer.movesSinceLastInc < 1 {
		return errors.New("cannot calculate score when no steps were made")
	}
	center := math.Abs(float64(scorer.OldHeadPos.X-scorer.OldFoodPos.X)) + math.Abs(float64(scorer.OldHeadPos.Y-scorer.OldFoodPos.Y))
	left := float64(scorer.OldHeadPos.X+scorer.gridWidth-scorer.OldFoodPos.X) + math.Abs(float64(scorer.OldHeadPos.Y-scorer.OldFoodPos.Y))
	right := float64(scorer.gridWidth-scorer.OldHeadPos.X+scorer.OldFoodPos.X) + math.Abs(float64(scorer.OldHeadPos.Y-scorer.OldFoodPos.Y))
	top := math.Abs(float64(scorer.OldHeadPos.X-scorer.OldFoodPos.X)) + float64(scorer.OldHeadPos.Y+scorer.gridHeight-scorer.OldFoodPos.Y)
	bottom := math.Abs(float64(scorer.OldHeadPos.X-scorer.OldFoodPos.X)) + float64(scorer.gridHeight-scorer.OldHeadPos.Y+scorer.OldFoodPos.Y)
	distance := math.Min(center, left)
	distance = math.Min(distance, right)
	distance = math.Min(distance, top)
	distance = math.Min(distance, bottom)
	// this should actually never happen
	if scorer.movesSinceLastInc < int(distance) {
		scorer.Score += scorer.maxPoints
		return nil
	}
	// too many steps -> 1 point
	if scorer.movesSinceLastInc-int(distance) >= (scorer.gridHeight*scorer.gridWidth)/2 {
		scorer.Score += 1
		return nil
	}
	x := (float64(scorer.movesSinceLastInc) - distance) / float64((scorer.gridHeight*scorer.gridWidth)/2)
	pointsFrac := (1 - math.Pow(2*x-1, 3)) / 2
	scorer.Score += int(math.Max(math.Round(pointsFrac*float64(scorer.maxPoints)), 1))
	return nil
}

func InitScorer(width, height int) Scorer {
	return Scorer{
		Score:             0,
		movesSinceLastInc: 0,
		weight:            50,
		gridWidth:         width,
		gridHeight:        height,
		maxPoints:         10,
	}
}

type Snail struct {
	// NOTE: the head is at the end of the slice!
	Body      []Pos
	Direction Velocity
	OldTail   Pos
}

func (snail *Snail) NextPos(oldPos Pos, XDim int, YDim int) Pos {
	newHead := Pos{X: 0, Y: 0}
	newHead.X = (oldPos.X + snail.Direction.X) % XDim
	newHead.Y = (oldPos.Y + snail.Direction.Y) % YDim
	if newHead.X < 0 {
		newHead.X = XDim - 1
	}
	if newHead.Y < 0 {
		newHead.Y = YDim - 1
	}
	return newHead
}

func (snail *Snail) MoveForward(ate bool, XDim int, YDim int) {
	oldHead := snail.Body[len(snail.Body)-1]
	newHead := snail.NextPos(oldHead, XDim, YDim)
	snail.Body = append(snail.Body, newHead)
	if !ate {
		snail.OldTail = snail.Body[0]
		snail.Body = snail.Body[1:]
	}
}

func (snail *Snail) GetHead() Pos {
	return snail.Body[len(snail.Body)-1]
}

func InitSnail(width int, height int) Snail {
	startPos := []Pos{{
		X: int(width / 2),
		Y: int(height / 2),
	},
		{
			X: int(width/2) + 1,
			Y: int(height / 2),
		},
		{
			X: int(width/2) + 2,
			Y: int(height / 2),
		},
	}
	oldTail := Pos{
		X: -1,
		Y: -1,
	}
	snail := Snail{
		Body:      startPos,
		Direction: EastDir,
		OldTail:   oldTail,
	}
	return snail
}

type Game struct {
	Food                  Pos
	Snail                 Snail
	Scorer                Scorer
	NextDirection         chan Velocity
	PauseChan             chan struct{}
	Paused                bool
	Screen                tcell.Screen
	GameDelayMilliSeconds time.Duration
	XDim                  int
	YDim                  int
	GameOver              bool
}

func InitScreen() tcell.Screen {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	defStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	screen.SetStyle(defStyle)
	return screen
}

func (game *Game) CreateFood() error {
	potentialFree := game.XDim*game.YDim - len(game.Snail.Body)
	if potentialFree < 1 {
		return errors.New("no free cell for food left")
	}
	next := rand.Intn(potentialFree)
	cur := 0
	for x := 0; x < game.XDim; x++ {
		for y := 0; y < game.YDim; y++ {
			toCheck := Pos{X: x, Y: y}
			if game.CheckCollisions(toCheck, game.Snail.Body) {
				continue
			}
			cur += 1
			if cur >= next {
				game.Food = toCheck
				return nil
			}
		}
	}
	return errors.New("no free cell for food left, unreachable")
}

func (game *Game) CheckCollisions(posToCheck Pos, potentialCollision []Pos) bool {
	for _, pos := range potentialCollision {
		if pos.X == posToCheck.X && pos.Y == posToCheck.Y {
			return true
		}
	}
	return false
}

func (game *Game) WonGame() bool {
	return game.XDim*game.YDim <= len(game.Snail.Body)
}

func (game *Game) AdjustDelay() {
	size := game.XDim * game.YDim
	share := (100 / float32(size)) * float32(len(game.Snail.Body))
	newDelay := game.GameDelayMilliSeconds.Milliseconds()
	if share > 90 && newDelay > 100 ||
		(share > 80 && newDelay > 110) ||
		(share > 70 && newDelay > 120) ||
		(share > 60 && newDelay > 130) ||
		(share > 50 && newDelay > 140) ||
		(share > 40 && newDelay > 150) ||
		(share > 30 && newDelay > 160) ||
		(share > 20 && newDelay > 170) ||
		(share > 10 && newDelay > 180) {
		newDelay -= 10
	}
	if newDelay < 100 {
		newDelay = 100
	}
	game.GameDelayMilliSeconds = time.Duration(newDelay) * time.Millisecond
}

func (game *Game) DrawBoard() {
	for c := 0; c < game.XDim+2; c++ {
		var ru = tcell.RuneHLine
		var rl = tcell.RuneHLine
		var double = true
		if c == 0 {
			double = false
			ru = tcell.RuneULCorner
			rl = tcell.RuneLLCorner
		} else if c == game.XDim+1 {
			double = false
			ru = tcell.RuneURCorner
			rl = tcell.RuneLRCorner
		}
		game.Screen.SetContent(c*2, 0, ru, nil, wallStyle)
		game.Screen.SetContent(c*2, game.YDim+1, rl, nil, wallStyle)
		if double {
			game.Screen.SetContent(c*2+1, 0, ru, nil, wallStyle)
			game.Screen.SetContent(c*2+1, game.YDim+1, rl, nil, wallStyle)
		}
	}
	game.Screen.SetContent(1, game.YDim+1, tcell.RuneHLine, nil, wallStyle)

	for r := 1; r < game.YDim+1; r++ {
		game.Screen.SetContent(0, r, tcell.RuneVLine, nil, wallStyle)
		game.Screen.SetContent(game.XDim*2+2, r, tcell.RuneVLine, nil, wallStyle)
	}

	game.Screen.SetContent(game.Food.X*2+1, game.Food.Y+1, tcell.RuneBlock, nil, foodStyle)
	game.Screen.SetContent(game.Food.X*2+2, game.Food.Y+1, tcell.RuneBlock, nil, foodStyle)
	for index, pos := range game.Snail.Body {
		var style = snailBodySytle
		if index == len(game.Snail.Body)-1 {
			style = snailHeadSytle
		}
		game.Screen.SetContent(pos.X*2+1, pos.Y+1, tcell.RuneBlock, nil, style)
		game.Screen.SetContent(pos.X*2+2, pos.Y+1, tcell.RuneBlock, nil, style)
	}

	score := fmt.Sprintf("Score: %d", game.Scorer.Score)
	for index, l := range score {
		game.Screen.SetContent(1+index, 0, l, nil, blackWhiteStyle)
	}
}

func (game *Game) DrawPause() {
	text := "Paused, wanna resume? p"
	row := game.YDim/2 + 1
	col := game.XDim/2 - len(text)/2 + 1
	for _, r := range text {
		game.Screen.SetContent(col, row, r, nil, blackWhiteStyle)
		col++
	}
}

func (game *Game) DrawGameOver(won bool) {
	first := "Game Over, you suck!"
	if won {
		first = "Game Over, you have WON!"
	}
	texts := [3]string{
		first,
		fmt.Sprintf("You reached a score of %d points.", game.Scorer.Score),
		"Play Again? y/n",
	}
	for index, text := range texts {
		row := game.YDim/2 + 1 + index
		col := game.XDim - len(text)/2 + 1
		for _, r := range text {
			game.Screen.SetContent(col, row, r, nil, blackWhiteStyle)
			col++
		}
	}
}

func (game *Game) IsValidNewDir(newDir Velocity) bool {
	if NorthDir.Equals(game.Snail.Direction) {
		return !SouthDir.Equals(newDir)
	} else if SouthDir.Equals(game.Snail.Direction) {
		return !NorthDir.Equals(newDir)
	} else if EastDir.Equals(game.Snail.Direction) {
		return !WestDir.Equals(newDir)
	} else if WestDir.Equals(game.Snail.Direction) {
		return !EastDir.Equals(newDir)
	}
	return NorthDir.Equals(newDir) || SouthDir.Equals(newDir) || WestDir.Equals(newDir) || EastDir.Equals(newDir)
}

func (game *Game) Loop(ctx context.Context) {
	game.ResetState()
	game.Scorer.OldHeadPos = game.Snail.GetHead()
	game.Scorer.OldFoodPos = game.Food
	for {
		select {
		case <-ctx.Done():
			// The context is over, stop processing results
			return
		case newDir := <-game.NextDirection:
			if game.IsValidNewDir(newDir) {
				game.Snail.Direction = newDir
			}
		case <-game.PauseChan:
			game.Paused = !game.Paused
			if game.Paused {
				select {
				case <-game.PauseChan:
					game.Paused = !game.Paused
				}
			}
		default:
			// dont block
		}
		var ate = false
		if game.CheckCollisions(game.Food, game.Snail.Body) {
			ate = true
			ErrExit(game.Scorer.CalculateScore())
			ErrExit(game.CreateFood())
			game.Scorer.OldHeadPos = game.Snail.GetHead()
			game.Scorer.OldFoodPos = game.Food
		}
		if game.CheckCollisions(game.Snail.Body[len(game.Snail.Body)-1], game.Snail.Body[:len(game.Snail.Body)-1]) {
			break
		}
		if game.WonGame() {
			break
		}
		game.Snail.MoveForward(ate, game.XDim, game.YDim)
		game.Scorer.Step()
		game.AdjustDelay()
		game.Screen.Clear()
		game.DrawBoard()
		game.Screen.Show()
		time.Sleep(game.GameDelayMilliSeconds)
	}
	game.GameOver = true
	game.DrawGameOver(game.WonGame())
	game.Screen.Show()
}

func (game *Game) CreateGameContext(ctx context.Context) (context.Context, context.CancelFunc) {
	toCancel, cancelFunc := context.WithCancel(ctx)
	return toCancel, cancelFunc
}

func (game *Game) Run(delayMilliseconds, dimensions int) {
	ctx := context.Background()
	var toCancel, cancelFunc = game.CreateGameContext(ctx)

	game.InitGame(delayMilliseconds, dimensions)
	go game.Loop(toCancel)

	for {
		switch event := game.Screen.PollEvent().(type) {
		case *tcell.EventResize:
			game.Screen.Sync()
		case *tcell.EventKey:
			if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyCtrlC {
				cancelFunc()
				game.Screen.Fini()
				return
			} else if event.Key() == tcell.KeyUp || event.Rune() == 'w' {
				game.NextDirection <- NorthDir
			} else if event.Key() == tcell.KeyDown || event.Rune() == 's' {
				game.NextDirection <- SouthDir
			} else if event.Key() == tcell.KeyLeft || event.Rune() == 'a' {
				game.NextDirection <- WestDir
			} else if event.Key() == tcell.KeyRight || event.Rune() == 'd' {
				game.NextDirection <- EastDir
			} else if event.Rune() == 'p' {
				dummy := struct{}{}
				game.PauseChan <- dummy
			} else if event.Rune() == 'y' && game.GameOver {
				cancelFunc()
				toCancel, cancelFunc = game.CreateGameContext(ctx)
				go game.Loop(toCancel)
			} else if event.Rune() == 'n' && game.GameOver {
				cancelFunc()
				game.Screen.Fini()
				return
			}
		}
	}
}

func (game *Game) UpdateDimesnions(dimension int) {
	//width, height := game.Screen.Size()
	game.Screen.SetSize(dimension+2, dimension+2)
	game.XDim = dimension
	game.YDim = dimension
}

func (game *Game) ResetState() {
	game.Snail = InitSnail(game.XDim, game.YDim)
	game.Scorer = InitScorer(game.XDim, game.YDim)
	ErrExit(game.CreateFood())
	game.GameOver = false
}

func (game *Game) InitGame(delayMilliseconds, dimensions int) {
	game.Screen = InitScreen()
	game.UpdateDimesnions(dimensions)
	game.ResetState()
	game.GameDelayMilliSeconds = time.Duration(delayMilliseconds) * time.Millisecond
	game.NextDirection = make(chan Velocity)
	game.PauseChan = make(chan struct{})
}

var Version = "development"

func main() {

	var gameDelayMilliSeconds = flag.Int("delay", 150,
		"starting delay in milliseconds of the game (min=100,max=200)")
	var dimensions = flag.Int("dimensions", 20, "x and y dimension of the game grid (min=10, max=50)")
	var printVersion = flag.Bool("version", false, "print version information")
	flag.Parse()

	if *printVersion {
		fmt.Printf("snail version %s\n", Version)
		os.Exit(0)
	}

	if *gameDelayMilliSeconds < 100 || *gameDelayMilliSeconds > 200 {
		*gameDelayMilliSeconds = 150
	}

	if *dimensions < 10 {
		*dimensions = 10
	} else if *dimensions > 50 {
		*dimensions = 50
	}

	game := Game{}
	game.Run(*gameDelayMilliSeconds, *dimensions)

	os.Exit(0)
}
