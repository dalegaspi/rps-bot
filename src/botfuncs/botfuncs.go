package botfuncs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"semaphore"
	"time"
)

type GameCell struct {
	Col        int
	IsOccupied bool
	PlayerId   int
	PlayerName string
	Row        int
	WinCount   int
}

type Player struct {
	Bombs           int
	BonusPoints     int
	Color           string
	EmailAddress    string
	ErrorAttacks    int
	LastActive      string
	Name            string
	NumberAttacks   int
	PaperAttacks    int
	PlayerId        int
	Points          int
	RockAttacks     int
	ScissorsAttacks int
	Score           int
	TeamId          int
	WonSquares      int
}

type GameBoard struct {
	Game struct {
		Cols                int
		CurrentPlayers      int
		End                 string
		GameCells           []GameCell
		GameType            string
		Id                  int
		Joined              bool
		LockedGame          bool
		PaperAttackLimit    int
		Players             []Player
		RockAttackLimit     int
		RollOverLimit       int
		Rows                int
		ScissorsAttackLimit int
		Start               string
		TotalCells          int
	}
	IsError bool
	Message string
}

type AttackResult struct {
	Clue       string
	GameCell   GameCell
	IsError    bool
	Message    string
	Player     Player
	Result     string
	ResultCode int
}

const baseGameUrl string = "http://10.1.105.226:806/GamePlay.svc"

var validWeapons = map[string]map[string]string{
	"rock":     {"conquers": "scissors", "weakness": "paper"},
	"paper":    {"conquers": "rock", "weakness": "scissors"},
	"scissors": {"conquers": "paper", "weakness": "rock"},
}

var cluesSem = semaphore.New(1)

// this is the cool part of how this is going to be implemented
// there will only be 18 clues that will tell you what is NOT
// the suspect, room, or weapon.  once there are 3 artifacts left
// you are now certain of solving the clue and they are in the order
// clues[0] = suspect, clues[1] = room, clues[2] = weapon
var clues = []string{
	// suspects

	"MissScarlet",
	"ColonelMustard",
	"MrsWhite",
	"ReverendGreen",
	"MrsPeacock",
	"ProfessorPlum",

	// rooms

	"Kitchen",
	"Ballroom",
	"Conservatory",
	"DiningRoom",
	"BilliardRoom",
	"Library",
	"Lounge",
	"Hall",
	"Study",

	// weapons

	"Candlestick",
	"Knife",
	"LeadPipe",
	"Revolver",
	"Rope",
	"Wrench",
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// This function fetch the content of a URL will return it as an
// array of bytes if retrieved successfully.
// http://www.codingcookies.com/2013/03/21/consuming-json-apis-with-go/
func getContent(url string) ([]byte, error) {
	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// Send the request via a client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// Defer the closing of the body
	defer resp.Body.Close()
	// Read the content into a byte array
	body, err := ioutil.ReadAll(resp.Body)
	//fmt.Println("YEEEE")
	if err != nil {
		return nil, err
	}
	// At this point we're done - simply return the bytes
	return body, nil
}

func GetGameInfo(gameId int, key string) (*GameBoard, error) {
	var url = fmt.Sprintf("%s/gameboard/%d/?playerAccessKey=%s", baseGameUrl, gameId, key)
	//	fmt.Printf("game info url %s\n", url)
	content, err := getContent(url)

	if err != nil {
		return nil, err
	}

	var board GameBoard
	err = json.Unmarshal(content, &board)

	if err != nil {
		return nil, err
	}

	return &board, err
}

func GetCellAtCoordinates(row int, col int, board *GameBoard) *GameCell {
	for i := 0; i < len(board.Game.GameCells); i++ {
		if board.Game.GameCells[i].Col == col && board.Game.GameCells[i].Row == row {
			return &board.Game.GameCells[i]
		}
	}

	return nil
}

func AttackCell(gameId int, row int, col int, weapon string, key string) (*AttackResult, error) {
	var url = fmt.Sprintf("%s/attack/%d/%d,%d/%s?playerAccessKey=%s", baseGameUrl, gameId, col, row, weapon, key)

	//fmt.Printf("attack url %s\n", url)
	content, err := getContent(url)

	if err != nil {
		return nil, err
	}

	var ar AttackResult
	err = json.Unmarshal(content, &ar)

	if err != nil {
		return nil, err
	}

	return &ar, err
}

func ClueSolve(gameId int, suspect string, room string, weapon string, key string) (*AttackResult, error) {
	var url = fmt.Sprintf("%s/clueSolve/%s/%s/%s?playerAccessKey=%s", baseGameUrl, suspect, room, weapon, key)

	fmt.Printf("clueSolve url %s\n", url)
	content, err := getContent(url)

	if err != nil {
		return nil, err
	}

	var ar AttackResult
	err = json.Unmarshal(content, &ar)

	if err != nil {
		return nil, err
	}

	return &ar, err
}

func GetRemainingClues(clue string) []string {
	cluesSem.Acquire()

	for i, v := range clues {
		if v == clue {
			clues = append(clues[:i], clues[i+1:]...)
		}
	}

	cluesSem.Release()

	return clues
}

func GetRandomWeapon() string {
	r := rand.Intn(len(validWeapons))
	i := 0
	var weapon string
	for k, _ := range validWeapons {
		if r == i {
			weapon = k
		}
		i++
	}

	return weapon
}

func GetWinningWeapon(weapon string) string {
	return validWeapons[weapon]["conquers"]
}

func GetLosingWeapon(weapon string) string {
	return validWeapons[weapon]["weakness"]
}

func GetRandomWeaponThatIsNot(weapon string) string {

	random_weapon := GetRandomWeapon()
	for random_weapon == weapon {
		random_weapon = GetRandomWeapon()
	}

	return random_weapon
}
