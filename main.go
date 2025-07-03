package main

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// RSS is the root of the RSS feed
type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	PubDate     string   `xml:"pubDate"`
	Categories  []string `xml:"category"`
}

// Comment represents a single comment in the Listview JSON
// Made global for reuse
type Comment struct {
	Body string `json:"body"`
	User string `json:"user"`
}

// GameComment ties a comment to its article index
// Used for the Connections-style game
type GameComment struct {
	Comment    Comment
	ArticleIdx int // Index of the article in the articles slice
}

// GameState holds the answer and used indices for a game
type GameState struct {
	Answer   []int         // article indices for each comment
	Used     map[int]bool  // which comment indices have been grouped
	Articles []Item        // articles for this game
	Comments []GameComment // original comments with quote counts
	Colors   []string      // color for each article
}

var (
	gameStore   = make(map[string]*GameState) // gameID -> GameState
	gameStoreMu sync.Mutex
)

// GameStartResponse is the response for /start-game
// ArticleIdx is omitted from comments for the client
type GameStartResponse struct {
	GameID   string         `json:"game_id"`
	Articles []Item         `json:"articles"`
	Comments []GameCommentC `json:"comments"`
}

type GameCommentC struct {
	Comment Comment `json:"comment"`
	Index   int     `json:"index"` // index in the shuffled list
}

// RemoveQuoteTags removes [quote=...]...[/quote] and [quote]...[/quote] tags, including nested ones, from a string
// It also removes empty lines from the result.
func RemoveQuoteTags(s string) string {
	for {
		start1 := strings.Index(s, "[quote=")
		start2 := strings.Index(s, "[quote]")
		var start int
		if start1 == -1 && start2 == -1 {
			break
		} else if start1 == -1 {
			start = start2
		} else if start2 == -1 {
			start = start1
		} else if start1 < start2 {
			start = start1
		} else {
			start = start2
		}
		end := strings.LastIndex(s[start:], "[/quote]")
		if end == -1 {
			break
		}
		end += start + len("[/quote]")
		s = s[:start] + s[end:]
	}
	// Remove empty lines
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

// Fetches 4 random articles, each with at least 4 comments, and 4 random comments per article
// Returns the articles and a shuffled slice of GameComment
func prepareConnectionsGame(rss *RSS) ([]Item, []GameComment, error) {
	// Filter articles with at least 4 comments
	qualified := []struct {
		item     Item
		comments []Comment
	}{}
	indices := rand.Perm(len(rss.Channel.Items))
	for _, idx := range indices {
		item := rss.Channel.Items[idx]
		resp, err := http.Get(item.Link)
		if err != nil {
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}
		re := regexp.MustCompile(`new Listview\(\{"id":"posts".*?\}\)`) // non-greedy match
		match := re.Find(body)
		if match == nil {
			continue
		}
		jsonStart := strings.Index(string(match), "(")
		jsonEnd := strings.LastIndex(string(match), ")")
		if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart+1 {
			continue
		}
		jsonStr := string(match)[jsonStart+1 : jsonEnd]
		type Listview struct {
			Data []Comment `json:"data"`
		}
		var lv Listview
		if err := json.Unmarshal([]byte(jsonStr), &lv); err != nil {
			continue
		}
		if len(lv.Data) >= 4 {
			qualified = append(qualified, struct {
				item     Item
				comments []Comment
			}{item, lv.Data})
			if len(qualified) == 4 {
				break
			}
		}
	}
	if len(qualified) < 4 {
		return nil, nil, fmt.Errorf("not enough articles with at least 4 comments")
	}
	articles := make([]Item, 4)
	gameComments := make([]GameComment, 0, 16)
	for i, q := range qualified {
		articles[i] = q.item
		indices := rand.Perm(len(q.comments))
		for j := 0; j < 4; j++ {
			c := q.comments[indices[j]]
			c.Body = RemoveQuoteTags(c.Body)
			gameComments = append(gameComments, GameComment{
				Comment:    c,
				ArticleIdx: i,
			})
		}
	}
	// Shuffle all 16 comments
	rand.Shuffle(len(gameComments), func(i, j int) {
		gameComments[i], gameComments[j] = gameComments[j], gameComments[i]
	})
	return articles, gameComments, nil
}

func generateGameID() string {
	b := make([]byte, 8)
	cryptorand.Read(b)
	return hex.EncodeToString(b)
}

func startGameHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("https://www.wowhead.com/news/rss/all")
	if err != nil {
		http.Error(w, "Failed to fetch RSS", 500)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read RSS", 500)
		return
	}
	var rss RSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		http.Error(w, "Failed to parse RSS", 500)
		return
	}
	articles, gameComments, err := prepareConnectionsGame(&rss)
	if err != nil {
		http.Error(w, "Failed to prepare game: "+err.Error(), 500)
		return
	}
	gameID := generateGameID()
	answer := make([]int, len(gameComments))
	commentsC := make([]GameCommentC, len(gameComments))
	for i, gc := range gameComments {
		answer[i] = gc.ArticleIdx
		commentsC[i] = GameCommentC{Comment: gc.Comment, Index: i}
	}
	// Calculate total quotes for each article
	articleQuoteCounts := make([]int, 4)
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			idx := i*4 + j
			articleQuoteCounts[i] += CountQuoteTags(gameComments[idx].Comment.Body)
		}
	}
	// Sort indices by quote count
	indices := []int{0, 1, 2, 3}
	sort.Slice(indices, func(i, j int) bool {
		return articleQuoteCounts[indices[i]] < articleQuoteCounts[indices[j]]
	})
	colors := make([]string, 4)
	nytColors := []string{"yellow", "green", "blue", "purple"}
	for rank, idx := range indices {
		colors[idx] = nytColors[rank]
	}
	gameStoreMu.Lock()
	gameStore[gameID] = &GameState{Answer: answer, Used: make(map[int]bool), Articles: articles, Comments: gameComments, Colors: colors}
	gameStoreMu.Unlock()

	// Debug: Print the correct solution when game is initiated
	fmt.Printf("Debug - New game %s:\n", gameID)
	for i := 0; i < 4; i++ {
		fmt.Printf("  Group %d: ", i+1)
		for j := 0; j < 4; j++ {
			idx := i*4 + j
			if idx < len(answer) {
				fmt.Printf("%d ", answer[idx])
			}
		}
		fmt.Printf("(Article %d)\n", i)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GameStartResponse{
		GameID:   gameID,
		Articles: articles,
		Comments: commentsC,
	})
}

type CheckSolutionRequest struct {
	GameID string `json:"game_id"`
	Group  []int  `json:"group"` // 4 comment indices
}

type CheckSolutionResponse struct {
	Correct      bool   `json:"correct"`
	Finished     bool   `json:"finished"`
	Remaining    int    `json:"remaining"`
	OneAway      bool   `json:"one_away"`
	ArticleTitle string `json:"article_title,omitempty"`
	ArticleURL   string `json:"article_url,omitempty"`
	Color        string `json:"color,omitempty"`
}

func checkSolutionHandler(w http.ResponseWriter, r *http.Request) {
	var req CheckSolutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", 400)
		return
	}
	gameStoreMu.Lock()
	gs, ok := gameStore[req.GameID]
	gameStoreMu.Unlock()
	if !ok {
		http.Error(w, "Game not found", 404)
		return
	}
	if len(req.Group) != 4 {
		http.Error(w, "Group must have 4 indices", 400)
		return
	}
	// Check if any index is already used or out of range
	for _, idx := range req.Group {
		if idx < 0 || idx >= len(gs.Answer) || gs.Used[idx] {
			http.Error(w, "Invalid or already used comment index", 400)
			return
		}
	}

	// Debug: Print the correct solution
	fmt.Printf("Debug - Game %s: Correct solution is %v\n", req.GameID, gs.Answer)

	// Check if all belong to the same article
	article := gs.Answer[req.Group[0]]
	correct := true
	for _, idx := range req.Group {
		if gs.Answer[idx] != article {
			correct = false
			break
		}
	}

	// Check if "one away" (3 out of 4 correct)
	oneAway := false
	if !correct {
		articleCounts := make(map[int]int)
		for _, idx := range req.Group {
			articleCounts[gs.Answer[idx]]++
		}
		for _, count := range articleCounts {
			if count == 3 {
				oneAway = true
				break
			}
		}
	}

	if correct {
		// Mark as used
		gameStoreMu.Lock()
		for _, idx := range req.Group {
			gs.Used[idx] = true
		}
		gameStoreMu.Unlock()
	}
	remaining := 0
	for i := range gs.Answer {
		if !gs.Used[i] {
			remaining++
		}
	}
	finished := remaining == 0

	// Get article info and color if correct
	var articleTitle, articleURL, color string
	if correct {
		articleTitle = gs.Articles[article].Title
		articleURL = gs.Articles[article].Link
		color = gs.Colors[article]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CheckSolutionResponse{
		Correct:      correct,
		Finished:     finished,
		Remaining:    remaining,
		OneAway:      oneAway,
		ArticleTitle: articleTitle,
		ArticleURL:   articleURL,
		Color:        color,
	})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		http.Error(w, "Template error", 500)
		return
	}
	tmpl.Execute(w, nil)
}

// CountQuoteTags counts the number of quote tags in a string
func CountQuoteTags(s string) int {
	count := 0
	for {
		start1 := strings.Index(s, "[quote=")
		start2 := strings.Index(s, "[quote]")
		var start int
		if start1 == -1 && start2 == -1 {
			break
		} else if start1 == -1 {
			start = start2
		} else if start2 == -1 {
			start = start1
		} else if start1 < start2 {
			start = start1
		} else {
			start = start2
		}
		end := strings.LastIndex(s[start:], "[/quote]")
		if end == -1 {
			break
		}
		end += start + len("[/quote]")
		s = s[:start] + s[end:]
		count++
	}
	return count
}

// GetDifficultyColor returns the NYT Connections color based on total quote count
func GetDifficultyColor(totalQuotes int) string {
	if totalQuotes <= 2 {
		return "yellow" // Easiest
	} else if totalQuotes <= 5 {
		return "green" // Medium
	} else if totalQuotes <= 8 {
		return "blue" // Harder
	} else {
		return "purple" // Hardest
	}
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/start-game", startGameHandler)
	http.HandleFunc("/check-solution", checkSolutionHandler)
	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
