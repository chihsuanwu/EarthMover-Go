// Command cli provides a terminal-based Gomoku game against the AI.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/todd/earthmover/internal/ai"
	"github.com/todd/earthmover/internal/board"
)

const dimen = board.Dimen

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Select rule
	fmt.Println("freestyle (1), renju-basic (2)")
	fmt.Print("> ")
	var ruleChoice int
	fmt.Fscan(reader, &ruleChoice)
	reader.ReadString('\n') // consume newline

	rule := board.RuleFreestyle
	if ruleChoice == 2 {
		rule = board.RuleRenjuBasic
	}

	// Select level
	fmt.Println("level: normal (0), advanced (1), master (2)")
	fmt.Print("> ")
	var level int
	fmt.Fscan(reader, &level)
	reader.ReadString('\n')

	if level < 0 || level > 2 {
		level = 0
	}

	a := ai.New()
	a.Reset(level, rule)

	// Display board state
	var stones [dimen][dimen]int // 0=empty, odd=black, even=white
	playNo := 0

	printBoard(stones)

	for {
		// AI thinks
		fmt.Println("AI searching...")
		start := time.Now()
		aiIdx := a.Think()
		elapsed := time.Since(start)
		fmt.Printf("time: %.3fs\n", elapsed.Seconds())

		if aiIdx == -1 {
			fmt.Println("AI passes")
		}

		// Start background thinking
		a.ThinkInBackground()

		// Get human input
		var row, col int
		for {
			fmt.Printf("enter move (A1 ~ %c%d): ", rune('A'+dimen-1), dimen)
			line, _ := reader.ReadString('\n')
			line = strings.TrimSpace(line)

			if parseInput(line, &row, &col) && stones[row][col] == 0 {
				break
			}
			fmt.Println("Invalid input")
		}

		// Stop background thinking
		a.StopBackground()

		// Play human move
		playNo++
		stones[row][col] = playNo
		printBoard(stones)

		winner := a.Play(row*dimen + col)
		if winner != -1 {
			if winner == 0 {
				fmt.Println("Black wins!")
			} else {
				fmt.Println("White wins!")
			}
			break
		}

		// Play AI move
		if aiIdx != -1 {
			aiRow := aiIdx / dimen
			aiCol := aiIdx % dimen
			playNo++
			stones[aiRow][aiCol] = playNo
			printBoard(stones)

			winner = a.Play(aiIdx)
			if winner != -1 {
				if winner == 0 {
					fmt.Println("Black wins!")
				} else {
					fmt.Println("White wins!")
				}
				break
			}
		}
	}
}

func parseInput(input string, row, col *int) bool {
	if len(input) < 2 || len(input) > 3 {
		return false
	}

	// Column: A-O or a-o
	c := input[0]
	if c >= 'A' && c < 'A'+dimen {
		*col = int(c - 'A')
	} else if c >= 'a' && c < 'a'+dimen {
		*col = int(c - 'a')
	} else {
		return false
	}

	// Row: 1-15
	numStr := input[1:]
	var r int
	for _, ch := range numStr {
		if ch < '0' || ch > '9' {
			return false
		}
		r = r*10 + int(ch-'0')
	}
	if r < 1 || r > dimen {
		return false
	}
	*row = r - 1
	return true
}

func printBoard(stones [dimen][dimen]int) {
	fmt.Println()
	// Column headers
	fmt.Print("    ")
	for c := 0; c < dimen; c++ {
		fmt.Printf("  %c ", rune('A'+c))
	}
	fmt.Println()

	for r := 0; r < dimen; r++ {
		// Row number
		fmt.Printf("%2d  ", r+1)

		for c := 0; c < dimen; c++ {
			if stones[r][c] != 0 {
				if stones[r][c]%2 == 1 {
					fmt.Print(" X  ")
				} else {
					fmt.Print(" O  ")
				}
			} else {
				cross := crossChar(r, c)
				// Each cell = 4 display columns:
				// First:  " " + cross + "──"  = 1+1+2 = 4
				// Middle: "─" + cross + "──"  = 1+1+2 = 4
				// Last:   "─" + cross + " "   = 1+1+2... no
				// Last:   "──" + cross + " "  = 2+1+1 = 4
				if c == 0 {
					fmt.Printf(" %s%s", cross, hRight(r))
				} else if c == dimen-1 {
					fmt.Printf("%s%s ", hLeft(r), cross)
				} else {
					fmt.Printf("%s%s%s", hLeft(r), cross, hRight(r))
				}
			}
		}
		fmt.Printf(" %d\n", r+1)

		// Vertical lines between rows
		if r < dimen-1 {
			fmt.Print("    ")
			for c := 0; c < dimen; c++ {
				if c == 0 || c == dimen-1 {
					fmt.Print(" ║  ")
				} else {
					fmt.Print(" │  ")
				}
			}
			fmt.Println()
		}
	}

	// Bottom column headers
	fmt.Print("    ")
	for c := 0; c < dimen; c++ {
		fmt.Printf("  %c ", rune('A'+c))
	}
	fmt.Println()
	fmt.Println()
}

// hLeft returns the horizontal line segment to the LEFT of a cross point.
// Returns single-width or double-width depending on position.
func hLeft(r int) string {
	if r == 0 || r == dimen-1 {
		return "═"
	}
	return "─"
}

// hRight returns the horizontal line segment to the RIGHT of a cross point (2 cols wide).
func hRight(r int) string {
	if r == 0 || r == dimen-1 {
		return "══"
	}
	return "──"
}

func crossChar(r, c int) string {
	isStarPoint := (r == 3 && c == 3) || (r == 3 && c == 11) ||
		(r == 11 && c == 3) || (r == 11 && c == 11) ||
		(r == 7 && c == 7)

	switch {
	case r == 0 && c == 0:
		return "╔"
	case r == 0 && c == dimen-1:
		return "╗"
	case r == dimen-1 && c == 0:
		return "╚"
	case r == dimen-1 && c == dimen-1:
		return "╝"
	case r == 0:
		return "╤"
	case r == dimen-1:
		return "╧"
	case c == 0:
		return "╟"
	case c == dimen-1:
		return "╢"
	case isStarPoint:
		return "╋"
	default:
		return "┼"
	}
}
