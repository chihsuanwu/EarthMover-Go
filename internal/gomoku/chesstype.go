// Package gomoku provides the shared data structures and logic
// for Gomoku board implementations (Freestyle and Renju).
package gomoku

// SingleType classifies a stone pattern along one direction for one color.
//   - Length: pattern length (0-5). 5 = five-in-a-row.
//   - Life: 0 = dead (blocked on one or both sides), 1 = live (open on both sides).
//   - Level: sub-classification for patterns of the same length and liveness.
type SingleType struct {
	Length int8
	Life   int8
	Level  int8
}

// Compare ordering: first by Length, then by Life.
// This matches the C++ operator< and operator> semantics.

func (t SingleType) Equal(o SingleType) bool {
	return t.Length == o.Length && t.Life == o.Life && t.Level == o.Level
}

func (t SingleType) Less(o SingleType) bool {
	if t.Length != o.Length {
		return t.Length < o.Length
	}
	return t.Life < o.Life
}

func (t SingleType) Greater(o SingleType) bool {
	if t.Length != o.Length {
		return t.Length > o.Length
	}
	return t.Life > o.Life
}

func (t SingleType) LessOrEqual(o SingleType) bool {
	return !t.Greater(o)
}

func (t SingleType) GreaterOrEqual(o SingleType) bool {
	return !t.Less(o)
}

// ChessType holds the pattern classification for both colors at a single point
// along one direction.
//   - Type[0] = Black's pattern
//   - Type[1] = White's pattern
type ChessType struct {
	Type [2]SingleType
}

func NewChessType(black, white SingleType) ChessType {
	return ChessType{Type: [2]SingleType{black, white}}
}

func (ct ChessType) Length(color int) int8  { return ct.Type[color].Length }
func (ct ChessType) Life(color int) int8    { return ct.Type[color].Life }
func (ct ChessType) Level(color int) int8   { return ct.Type[color].Level }

func (ct ChessType) Equal(o ChessType) bool {
	return ct.Type[0].Equal(o.Type[0]) && ct.Type[1].Equal(o.Type[1])
}
