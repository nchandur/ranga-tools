package pgn

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/nchandur/ranga-tools/internal/board"
)

// marks a game with a missing or "*" result
const ResultUnknown float32 = -1

var (
	tagRe        = regexp.MustCompile(`(?m)^[ \t]*\[(\w+)\s+"(.*)"\][ \t\r]*$`)
	eventTagRe   = regexp.MustCompile(`(?m)^\[Event\s`)
	moveNumRe    = regexp.MustCompile(`^\d+\.+`)
	resultScores = map[string]float32{"1-0": 1, "0-1": 0, "1/2-1/2": 0.5}
)

type Game struct {
	White   string
	Black   string
	Result  float32
	FEN     string
	Variant string
	Moves   []string
	SAN     []string
	board.Board
}

// returns the header tags and the movetext that follows them
func parseTags(pgn string) (tags map[string]string, movetext string) {
	tags, movetext = map[string]string{}, pgn
	for _, m := range tagRe.FindAllStringSubmatchIndex(pgn, -1) {
		tags[pgn[m[2]:m[3]]] = pgn[m[4]:m[5]]
		movetext = pgn[m[1]:]
	}
	return
}

// returns the White and Black tags of a raw game
func Players(pgn string) (white, black string) {
	tags, _ := parseTags(pgn)
	return tags["White"], tags["Black"]
}

func NewGame(pgn string) (Game, error) {
	tags, movetext := parseTags(pgn)
	g := Game{
		White:   tags["White"],
		Black:   tags["Black"],
		Variant: strings.TrimSpace(tags["Variant"]),
		Result:  ResultUnknown,
		FEN:     board.START,
		Board:   board.NewBoard(),
	}
	if r, ok := resultScores[tags["Result"]]; ok {
		g.Result = r
	}
	if fen := tags["FEN"]; fen != "" {
		g.FEN = fen
	}
	if err := g.Board.ParseFEN(g.FEN); err != nil {
		return Game{}, err
	}
	if err := g.parseMoves(movetext); err != nil {
		return Game{}, err
	}
	return g, nil
}

func (g *Game) parseMoves(movetext string) error {
	sans, err := tokenize(movetext)
	if err != nil {
		return fmt.Errorf("movetext: %w", err)
	}
	g.Moves = make([]string, 0, len(sans))
	g.SAN = sans

	b := g.Board.Preserve()
	for i, san := range sans {
		m := ParseSAN(&b, san)
		if m == board.NOMOVE || !b.MakeMove(m, false) {
			return fmt.Errorf("ply %d (%q) is illegal or unparseable in position %s", i+1, san, b.FEN())
		}
		b.Ply++
		g.Moves = append(g.Moves, m.String())
	}
	return nil
}

// tokenize returns the SAN moves of PGN movetext
func tokenize(s string) ([]string, error) {
	var moves []string
	depth := 0
	for i := 0; i < len(s); {
		switch c := s[i]; {
		case c == '{':
			end := strings.IndexByte(s[i:], '}')
			if end < 0 {
				return nil, fmt.Errorf("unterminated comment at byte %d", i)
			}
			i += end + 1
		case c == '(':
			depth++
			i++
		case c == ')':
			depth = max(depth-1, 0)
			i++
		case depth > 0 || strings.IndexByte(" \t\r\n", c) >= 0:
			i++
		default:
			j := i
			for j < len(s) && strings.IndexByte(" \t\r\n{(", s[j]) < 0 {
				j++
			}
			word := strings.TrimRight(moveNumRe.ReplaceAllString(s[i:j], ""), "!?")
			i = j
			switch word {
			case "", "1-0", "0-1", "1/2-1/2", "*":
			default:
				if word[0] != '$' {
					moves = append(moves, word)
				}
			}
		}
	}
	return moves, nil
}

// reports whether the game is ordinary chess
func (g *Game) IsStandard() bool {
	switch strings.ToLower(g.Variant) {
	case "", "standard", "chess", "from position":
		return true
	}
	return false
}

// splits concatenated PGN games
func SplitGames(data string) []string {
	data = strings.TrimPrefix(data, "\ufeff")
	idx := eventTagRe.FindAllStringIndex(data, -1)
	if len(idx) == 0 {
		if t := strings.TrimSpace(data); t != "" {
			return []string{t}
		}
		return nil
	}
	var games []string
	for i, m := range idx {
		end := len(data)
		if i+1 < len(idx) {
			end = idx[i+1][0]
		}
		if g := strings.TrimSpace(data[m[0]:end]); g != "" {
			games = append(games, g)
		}
	}
	return games
}

// ReadGames returns the raw text of every game in a PGN file.
func ReadGames(source string) ([]string, error) {
	data, err := os.ReadFile(source)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	games := SplitGames(string(data))
	if len(games) == 0 {
		return nil, fmt.Errorf("no game found in %s", source)
	}
	return games, nil
}
