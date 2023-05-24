// bot project main.go
package main

import (
	"botfuncs"
	"flag"
	"fmt"
	"os"
	"semaphore"
	"strings"
	"sync"
	"time"
)

/*
const attackerKey = "eNoBHwDg%2fzZ8fCoqfHxCbGFrZUxpdmVseXx8Kip8fHx8Kip8fDC4JguH"
const attacker = "BlakeLively"


var attackerKey string = "eNoBGQDm%2fzIwfHwqKnx8bG9senx8Kip8fHx8Kip8fDB1NQkg"
var attacker string = "lolz"
*/

var attackerKey = "eNoBIwDc%2fzI1fHwqKnx8Y2FybHlSYWVKZXBzZW58fCoqfHx8fCoqfHww58EM%2fA%3d%3d"
var attacker = "carlyRaeJepsen"
var gameId = 1

/*
const attackerKey = "eNoBIADf%2fzh8fCoqfHxSeWFuUmV5bm9sZHN8fCoqfHx8fCoqfHwwyP0MHw%3d%3d"
const attacker = "RyanReynolds"
*/
var protectorKey = "eNoBIADf%2fzh8fCoqfHxSeWFuUmV5bm9sZHN8fCoqfHx8fCoqfHwwyP0MHw%3d%3d"
var protector = "RyanReynolds"

var protect = false

var boardSem *semaphore.Semaphore = semaphore.New(1)
var currentBoard *botfuncs.GameBoard

type relayData struct {
	Row    int
	Col    int
	Weapon string
}

func updateBoard(gameId int, key string) {
	// TODO: add semaphores
	for {
		boardSem.Acquire()
		currentBoard, _ = botfuncs.GetGameInfo(gameId, key)
		boardSem.Release()
		time.Sleep(100 * time.Millisecond)
		fmt.Println("board data updated")
	}
}

func getCellToAttack(row_seed int, col_seed int) (int, int, string) {

	r := row_seed
	c := col_seed
	defer boardSem.Release()
	boardSem.Acquire()
	for ; r < currentBoard.Game.Rows; r++ {
		for ; c < currentBoard.Game.Cols; c++ {
			cell := botfuncs.GetCellAtCoordinates(r, c, currentBoard)
			/*
				if cell != nil && cell.PlayerName == attacker && currentBoard.Game.GameType == "NoHolesBarred" {
					var ar, _ = botfuncs.AttackCell(gameId, cell.Row, cell.Col, "broccoli", attackerKey)
					fmt.Printf("[%d,%d] %s result is %s: %s\n", cell.Row, cell.Col, "broccoli", ar.Result, ar.Message)
				}
			*/
			if cell != nil && currentBoard.Game.GameType == "Clue" {
				return cell.Row, cell.Col, cell.PlayerName
			} else if cell != nil && cell.PlayerName != attacker &&
				(cell.WinCount < currentBoard.Game.RollOverLimit ||
					//currentBoard.Game.GameType == "NoHolesBarred" ||
					currentBoard.Game.GameType == "BombsAway") {
				//boardSem.Release()
				return cell.Row, cell.Col, cell.PlayerName
			}
		}
		c = 0
	}

	//boardSem.Release()
	return -1, -1, ""
}

func updateCellOwnerAtCoordinates(row int, col int, new_owner string) {
	boardSem.Acquire()
	cell := botfuncs.GetCellAtCoordinates(row, col, currentBoard)
	cell.PlayerName = new_owner
	boardSem.Release()
}

func occupyCell(gameId int, row int, col int, limit int, wg *sync.WaitGroup, throttle chan int) {
	defer wg.Done()

	w := botfuncs.GetRandomWeapon()

	var cell *botfuncs.GameCell
	for lc := 0; lc < limit; lc++ {
		ar, err := botfuncs.AttackCell(gameId, row, col, w, attackerKey)

		if err != nil {
			break
		}

		cell = &ar.GameCell
		if cell != nil && (cell.WinCount >= currentBoard.Game.RollOverLimit ||
			//currentBoard.Game.GameType == "NoHolesBarred" ||
			currentBoard.Game.GameType == "BombsAway") {
			fmt.Printf("[%d,%d] already above winCount limit %d \n", row, col, cell.WinCount)

		}
		//fmt.Printf("[%d,%d] %s result is %s: %s\n", row, col, w, ar.Result, ar.Message)
		if ar.Result == "tie" {
			w = botfuncs.GetLosingWeapon(w)
		} else if ar.Result == "loss" {
			w = botfuncs.GetWinningWeapon(botfuncs.GetWinningWeapon(w))
		} else if ar.Result == "failure" {
			break
		} else {

			updateCellOwnerAtCoordinates(row, col, ar.Player.Name)

			// TODO:
			// if this is "Clue" try to solve if there are only 3 clues left
			if len(ar.Clue) > 0 {
				fmt.Printf("clue found at [%d:%d] : %s\n", row, col, ar.Clue)

				clues_split := strings.Split(ar.Clue, " ")
				clue := strings.Replace(clues_split[len(clues_split)-1], "\n", "", -1)

				fmt.Printf("actual clue at [%d:%d] : %s\n", row, col, clue)

				clues := botfuncs.GetRemainingClues(clue)

				if len(clues) == 3 {
					fmt.Printf("solving clue: %s %s %s\n", clues[0], clues[1], clues[2])
					// TODO: clueSolve
					os.Exit(0)
				} else {
					fmt.Printf("cannot solve clue yet; there are %d remaining clues\n", len(clues))
				}
			} else {
				fmt.Printf("no clue found at [%d:%d]\n", row, col)
			}
		}

	}

	<-throttle
}

func main() {
	//board, _ := botfuncs.GetGameInfo(2, attackerKey)
	//fmt.Printf("board message is [%s]\n", board.Message)

	//ar, _ := botfuncs.AttackCell(2, 0, 1, "scissors", attackerKey)
	//fmt.Printf("attack result is [%s]\n", ar.Result)

	flag.IntVar(&gameId, "g", 2, "game Id")
	flag.StringVar(&attackerKey, "k", "eNoBIwDc%2fzI1fHwqKnx8Y2FybHlSYWVKZXBzZW58fCoqfHx8fCoqfHww58EM%2fA%3d%3d", "player key")
	flag.StringVar(&attacker, "n", "carlyRaeJepsen", "player name")
	flag.StringVar(&protectorKey, "p", "eNoBIADf%2fzh8fCoqfHxSeWFuUmV5bm9sZHN8fCoqfHx8fCoqfHwwyP0MHw%3d%3d", "protector key")
	flag.BoolVar(&protect, "t", false, "protect")

	flag.Parse()

	currentBoard, _ = botfuncs.GetGameInfo(gameId, attackerKey)

	go updateBoard(gameId, attackerKey)

	r := 0
	c := 0
	var p string

	fmt.Printf("playing game [%d] with attacker [%s] and key [%s]\n", gameId, attacker, attackerKey)

	// throttling

	const maxConcurrency = 128

	var throttle = make(chan int, maxConcurrency)
	//var relay = make(chan relayData)

	var wg sync.WaitGroup

	for {
		if r >= 0 && c >= 0 {
			r, c, p = getCellToAttack(r, c)
			if r < 0 || c < 0 {
				fmt.Printf("no cell to attack")
				r = 0
				c = 0
				continue
			} else {
				fmt.Printf("cell to attack [%d:%d] owned by %s \n", r, c, p)
			}
		}

		throttle <- 1
		wg.Add(1)
		go occupyCell(gameId, r, c, 4, &wg, throttle)

		c++
		if c >= currentBoard.Game.Cols {
			c = 0
			r++
		}

		if r >= currentBoard.Game.Rows {
			r = 0
		}
	}

	wg.Wait()

}
