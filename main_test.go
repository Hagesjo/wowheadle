package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeArticle struct {
	Title string
	Link  string
}

type fakeComment struct {
	Body string
	User string
}

func TestPlayEntireGame(t *testing.T) {
	// Setup fake data
	// articles := []Item{
	// 	{Title: "A1", Link: "L1"},
	// 	{Title: "A2", Link: "L2"},
	// 	{Title: "A3", Link: "L3"},
	// 	{Title: "A4", Link: "L4"},
	// }
	comments := []GameComment{
		{Comment: Comment{Body: "C1", User: "U1"}, ArticleIdx: 0},
		{Comment: Comment{Body: "C2", User: "U2"}, ArticleIdx: 0},
		{Comment: Comment{Body: "C3", User: "U3"}, ArticleIdx: 0},
		{Comment: Comment{Body: "C4", User: "U4"}, ArticleIdx: 0},
		{Comment: Comment{Body: "C5", User: "U5"}, ArticleIdx: 1},
		{Comment: Comment{Body: "C6", User: "U6"}, ArticleIdx: 1},
		{Comment: Comment{Body: "C7", User: "U7"}, ArticleIdx: 1},
		{Comment: Comment{Body: "C8", User: "U8"}, ArticleIdx: 1},
		{Comment: Comment{Body: "C9", User: "U9"}, ArticleIdx: 2},
		{Comment: Comment{Body: "C10", User: "U10"}, ArticleIdx: 2},
		{Comment: Comment{Body: "C11", User: "U11"}, ArticleIdx: 2},
		{Comment: Comment{Body: "C12", User: "U12"}, ArticleIdx: 2},
		{Comment: Comment{Body: "C13", User: "U13"}, ArticleIdx: 3},
		{Comment: Comment{Body: "C14", User: "U14"}, ArticleIdx: 3},
		{Comment: Comment{Body: "C15", User: "U15"}, ArticleIdx: 3},
		{Comment: Comment{Body: "C16", User: "U16"}, ArticleIdx: 3},
	}
	// Insert fake game state
	gameID := "testgameid"
	answer := make([]int, len(comments))
	for i, gc := range comments {
		answer[i] = gc.ArticleIdx
	}
	gameStoreMu.Lock()
	gameStore[gameID] = &GameState{Answer: answer, Used: make(map[int]bool)}
	gameStoreMu.Unlock()

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(checkSolutionHandler))
	defer ts.Close()

	// Play the game: submit 4 correct groups
	groups := [][]int{
		{0, 1, 2, 3},     // All ArticleIdx 0
		{4, 5, 6, 7},     // All ArticleIdx 1
		{8, 9, 10, 11},   // All ArticleIdx 2
		{12, 13, 14, 15}, // All ArticleIdx 3
	}
	for i, group := range groups {
		reqBody, _ := json.Marshal(CheckSolutionRequest{
			GameID: gameID,
			Group:  group,
		})
		resp, err := http.Post(ts.URL, "application/json", bytes.NewReader(reqBody))
		if err != nil {
			t.Fatalf("failed to POST: %v", err)
		}
		var res CheckSolutionResponse
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			resp.Body.Close()
			t.Fatalf("failed to decode response: %v", err)
		}
		resp.Body.Close()
		if !res.Correct {
			t.Errorf("group %d should be correct", i)
		}
		if i < 3 && res.Finished {
			t.Errorf("game should not be finished after group %d", i)
		}
		if i == 3 && !res.Finished {
			t.Errorf("game should be finished after last group")
		}
	}
}

func TestWrongAnswer(t *testing.T) {
	// Setup fake data
	comments := []GameComment{
		{Comment: Comment{Body: "C1", User: "U1"}, ArticleIdx: 0},
		{Comment: Comment{Body: "C2", User: "U2"}, ArticleIdx: 0},
		{Comment: Comment{Body: "C3", User: "U3"}, ArticleIdx: 0},
		{Comment: Comment{Body: "C4", User: "U4"}, ArticleIdx: 0},
		{Comment: Comment{Body: "C5", User: "U5"}, ArticleIdx: 1},
		{Comment: Comment{Body: "C6", User: "U6"}, ArticleIdx: 1},
		{Comment: Comment{Body: "C7", User: "U7"}, ArticleIdx: 1},
		{Comment: Comment{Body: "C8", User: "U8"}, ArticleIdx: 1},
		{Comment: Comment{Body: "C9", User: "U9"}, ArticleIdx: 2},
		{Comment: Comment{Body: "C10", User: "U10"}, ArticleIdx: 2},
		{Comment: Comment{Body: "C11", User: "U11"}, ArticleIdx: 2},
		{Comment: Comment{Body: "C12", User: "U12"}, ArticleIdx: 2},
		{Comment: Comment{Body: "C13", User: "U13"}, ArticleIdx: 3},
		{Comment: Comment{Body: "C14", User: "U14"}, ArticleIdx: 3},
		{Comment: Comment{Body: "C15", User: "U15"}, ArticleIdx: 3},
		{Comment: Comment{Body: "C16", User: "U16"}, ArticleIdx: 3},
	}
	// Insert fake game state
	gameID := "testgameid2"
	answer := make([]int, len(comments))
	for i, gc := range comments {
		answer[i] = gc.ArticleIdx
	}
	gameStoreMu.Lock()
	gameStore[gameID] = &GameState{Answer: answer, Used: make(map[int]bool)}
	gameStoreMu.Unlock()

	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(checkSolutionHandler))
	defer ts.Close()

	// Submit a wrong group (mixing comments from different articles)
	wrongGroup := []int{0, 1, 4, 5} // Mix of ArticleIdx 0 and 1
	reqBody, _ := json.Marshal(CheckSolutionRequest{
		GameID: gameID,
		Group:  wrongGroup,
	})
	resp, err := http.Post(ts.URL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("failed to POST: %v", err)
	}
	var res CheckSolutionResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		resp.Body.Close()
		t.Fatalf("failed to decode response: %v", err)
	}
	resp.Body.Close()
	if res.Correct {
		t.Error("wrong group should not be correct")
	}
	if res.Finished {
		t.Error("game should not be finished after wrong answer")
	}
	if res.Remaining != 16 {
		t.Errorf("expected 16 remaining, got %d", res.Remaining)
	}
}
