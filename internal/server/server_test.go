package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func postJSON(handler http.HandlerFunc, body any) *httptest.ResponseRecorder {
	data, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func TestStartAndThink(t *testing.T) {
	srv := New("/tmp") // static dir doesn't matter for API tests

	// Start a game
	w := postJSON(srv.handleStart, map[string]any{
		"sessionId": "test-session-1",
		"level":     0,
		"rule":      0, // freestyle
	})
	if w.Code != http.StatusNoContent {
		t.Fatalf("/start: status = %d, want 204", w.Code)
	}

	// Think
	w = postJSON(srv.handleThink, map[string]any{
		"sessionId": "test-session-1",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("/think: status = %d, want 200", w.Code)
	}

	var thinkResp map[string]any
	json.NewDecoder(w.Body).Decode(&thinkResp)
	if thinkResp["row"] == nil || thinkResp["col"] == nil {
		t.Error("/think: missing row/col in response")
	}
	t.Logf("/think response: %v", thinkResp)
}

func TestPlayAndWin(t *testing.T) {
	srv := New("/tmp")

	// Start
	postJSON(srv.handleStart, map[string]any{
		"sessionId": "test-win",
		"level":     0,
		"rule":      0,
	})

	// Play moves to make black win (5 in a row on row 7)
	moves := [][2]int{
		{7, 3}, {0, 0}, // black, white
		{7, 4}, {0, 1},
		{7, 5}, {0, 2},
		{7, 6}, {0, 3},
		{7, 7}, // black wins
	}

	for i, m := range moves {
		w := postJSON(srv.handlePlay, map[string]any{
			"sessionId": "test-win",
			"row":       m[0],
			"col":       m[1],
		})
		if w.Code != http.StatusOK {
			t.Fatalf("move %d: status = %d", i, w.Code)
		}

		var resp map[string]any
		json.NewDecoder(w.Body).Decode(&resp)
		winner := int(resp["winner"].(float64))

		if i < len(moves)-1 {
			if winner != -1 {
				t.Fatalf("move %d: unexpected winner %v", i, winner)
			}
		} else {
			if winner != 0 {
				t.Errorf("last move: winner = %v, want 0 (black)", winner)
			}
		}
	}
}

func TestInvalidSession(t *testing.T) {
	srv := New("/tmp")

	w := postJSON(srv.handlePlay, map[string]any{
		"sessionId": "nonexistent",
		"row":       7,
		"col":       7,
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("/play invalid session: status = %d, want 400", w.Code)
	}
}

func TestKeepAlive(t *testing.T) {
	srv := New("/tmp")

	postJSON(srv.handleStart, map[string]any{
		"sessionId": "keepalive-test",
		"level":     0,
		"rule":      0,
	})

	w := postJSON(srv.handleKeepAlive, map[string]any{
		"sessionId": "keepalive-test",
	})
	if w.Code != http.StatusNoContent {
		t.Errorf("/keepAlive: status = %d, want 204", w.Code)
	}
}

func TestQuit(t *testing.T) {
	srv := New("/tmp")

	postJSON(srv.handleStart, map[string]any{
		"sessionId": "quit-test",
		"level":     0,
		"rule":      0,
	})

	w := postJSON(srv.handleQuit, map[string]any{
		"sessionId": "quit-test",
	})
	if w.Code != http.StatusNoContent {
		t.Errorf("/quit: status = %d, want 204", w.Code)
	}

	// Session should be gone
	w = postJSON(srv.handlePlay, map[string]any{
		"sessionId": "quit-test",
		"row":       7,
		"col":       7,
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("after quit, /play: status = %d, want 400", w.Code)
	}
}

func TestMaxInstances(t *testing.T) {
	srv := New("/tmp")

	// Fill all 3 slots
	for i := 0; i < maxInstances; i++ {
		w := postJSON(srv.handleStart, map[string]any{
			"sessionId": fmt.Sprintf("session-%d", i),
			"level":     0,
			"rule":      0,
		})
		if w.Code != http.StatusNoContent {
			t.Fatalf("start %d: status = %d", i, w.Code)
		}
	}

	// 4th should fail
	w := postJSON(srv.handleStart, map[string]any{
		"sessionId": "session-overflow",
		"level":     0,
		"rule":      0,
	})
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("4th start: status = %d, want 503", w.Code)
	}
}

func TestUndo(t *testing.T) {
	srv := New("/tmp")

	postJSON(srv.handleStart, map[string]any{
		"sessionId": "undo-test",
		"level":     0,
		"rule":      0,
	})

	// Play a move
	postJSON(srv.handlePlay, map[string]any{
		"sessionId": "undo-test",
		"row":       7,
		"col":       7,
	})

	// Undo it
	w := postJSON(srv.handleUndo, map[string]any{
		"sessionId": "undo-test",
		"times":     1,
	})
	if w.Code != http.StatusNoContent {
		t.Errorf("/undo: status = %d, want 204", w.Code)
	}
}
