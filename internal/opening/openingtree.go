// Package opening provides the opening book for the first few moves of a Gomoku game.
package opening

import (
	"bufio"
	_ "embed"
	"math/rand/v2"
	"strings"
	"sync"

	"github.com/todd/earthmover/internal/board"
)

//go:embed opening.txt
var openingData string

// node is a trie node for the opening tree.
// childNode is indexed by [rowOffset][colOffset][color],
// where offsets are relative to the origin (lower-right occupied cell).
type node struct {
	children [5][5][2]*node
	results  [][2]int // list of [rowOffset, colOffset] for suggested moves
}

var (
	root     *node
	initOnce sync.Once
)

// Init initializes the opening tree from the embedded data file.
func Init() {
	initOnce.Do(func() {
		root = &node{}
		parseAndBuild()
	})
}

// PointStatus provides read access to a board point's stone status.
type PointStatus interface {
	Status() board.StoneStatus
}

// Classify looks up the current board state in the opening tree.
// Returns a board index (0-224) for the suggested move, or -1 if no match.
func Classify(points []PointStatus) int {
	if root == nil {
		return -1
	}

	currentNode := root

	// Find bounding box of occupied points
	oriR, oriC := 0, 0
	left, top := 14, 14

	for r := 0; r < board.Dimen; r++ {
		for c := 0; c < board.Dimen; c++ {
			s := points[r*board.Dimen+c].Status()
			if s == board.Black || s == board.White {
				if r < top {
					top = r
				}
				if c < left {
					left = c
				}
				if r > oriR {
					oriR = r
				}
				if c > oriC {
					oriC = c
				}
			}
		}
	}

	// If the pattern is larger than 5x5, no match
	if oriR-top > 4 || oriC-left > 4 {
		return -1
	}

	// Walk the trie from origin (lower-right) scanning right-to-left, bottom-to-top
	curR, curC := oriR, oriC
	for {
		idx := curR*board.Dimen + curC
		s := points[idx].Status()
		if s == board.Black || s == board.White {
			color := int(s) // 0=black, 1=white
			rowOff := oriR - curR
			colOff := oriC - curC
			if currentNode.children[rowOff][colOff][color] == nil {
				return -1
			}
			currentNode = currentNode.children[rowOff][colOff][color]
		}

		curC--
		if curC < oriC-4 || curC < 0 {
			if curR == oriR-4 || curR == 0 {
				break
			}
			curR--
			curC = oriC
		}
	}

	// Pick a random result that falls within the valid center region
	index := -1
	count := 1
	for _, result := range currentNode.results {
		r := oriR - result[0]
		c := oriC - result[1]
		if r < 4 || r > 10 || c < 4 || c > 10 {
			continue
		}
		if rand.Float64() <= 1.0/float64(count) {
			index = r*board.Dimen + c
		}
		count++
	}
	return index
}

func parseAndBuild() {
	scanner := bufio.NewScanner(strings.NewReader(openingData))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Check if this line is a number label (like "2.1", "4.13")
		if len(line) > 0 && line[0] >= '0' && line[0] <= '9' {
			// Read the 5x5 table
			var table [5][5]byte
			for r := 0; r < 5; r++ {
				if !scanner.Scan() {
					return
				}
				row := scanner.Text()
				for c := 0; c < 5 && c < len(row); c++ {
					table[r][c] = row[c]
				}
			}

			// Insert in 8 orientations: rotate 4 times, mirror, rotate 4 times
			for m := 0; m < 2; m++ {
				for r := 0; r < 4; r++ {
					insert(table)
					table = rotate(table)
				}
				table = mirror(table)
			}
		}
	}
}

// rotate90 rotates the table 90 degrees clockwise.
func rotate(table [5][5]byte) [5][5]byte {
	var result [5][5]byte
	for r := 0; r < 5; r++ {
		for c := 0; c < 5; c++ {
			result[c][4-r] = table[r][c]
		}
	}
	return result
}

// mirror transposes the table (swap rows and columns).
func mirror(table [5][5]byte) [5][5]byte {
	var result [5][5]byte
	for r := 0; r < 5; r++ {
		for c := 0; c < 5; c++ {
			result[c][r] = table[r][c]
		}
	}
	return result
}

func insert(table [5][5]byte) {
	currentNode := root

	// Find origin: lower-right occupied cell
	oriR, oriC := 0, 0
	for r := 0; r < 5; r++ {
		for c := 0; c < 5; c++ {
			if table[r][c] == 'X' || table[r][c] == 'O' {
				if r > oriR {
					oriR = r
				}
				if c > oriC {
					oriC = c
				}
			}
		}
	}

	// Walk from origin, right-to-left, bottom-to-top
	curR, curC := oriR, oriC
	for {
		if table[curR][curC] == 'X' || table[curR][curC] == 'O' {
			color := 0
			if table[curR][curC] == 'O' {
				color = 1
			}
			rowOff := oriR - curR
			colOff := oriC - curC
			if currentNode.children[rowOff][colOff][color] == nil {
				currentNode.children[rowOff][colOff][color] = &node{}
			}
			currentNode = currentNode.children[rowOff][colOff][color]
		}

		curC--
		if curC < 0 {
			if curR == 0 {
				break
			}
			curR--
			curC = 4
		}
	}

	// Record suggested moves ('P')
	for curR := 0; curR < 5; curR++ {
		for curC := 0; curC < 5; curC++ {
			if table[curR][curC] == 'P' {
				result := [2]int{oriR - curR, oriC - curC}
				// Avoid duplicates
				found := false
				for _, r := range currentNode.results {
					if r == result {
						found = true
						break
					}
				}
				if !found {
					currentNode.results = append(currentNode.results, result)
				}
			}
		}
	}
}
