package pgn

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/nchandur/ranga-tools/internal/board"
)

var whiteRe = regexp.MustCompile(`\[White\s*"(.+)"\]`)
var blackRe = regexp.MustCompile(`\[Black\s*"(.+)"\]`)
var resultRe = regexp.MustCompile(`\[Result\s*"(.+)"\]`)
var fenRe = regexp.MustCompile(`\[FEN\s*"(.+)"\]`)
var taglineRe = regexp.MustCompile(`(?m)^\s*\[.*\]\s*$`)
var scoreRe = regexp.MustCompile(`\[%eval\s+([+-])?(?:M([0-9]+)|([0-9]+(?:\.[0-9]+)?))\]`)

// marks the start of a new game header block
var eventTagPattern = regexp.MustCompile(`(?m)^\[Event\s`)

type Game struct {
	White  string
	Black  string
	Result float32
	FEN    string
	Moves  []string
	board.Board
}

func NewGame(pgn string) (Game, error) {

	var game Game
	game.Board = board.NewBoard()
	game.parsePlayers(pgn)
	game.parseResult(pgn)
	game.parseFEN(pgn)

	if err := game.Board.ParseFEN(game.FEN); err != nil {
		return Game{}, err
	}

	game.parseMoves(pgn)

	return game, nil
}

// parse players
func (g *Game) parsePlayers(pgn string) {
	if matches := whiteRe.FindStringSubmatch(pgn); len(matches) > 0 {
		g.White = matches[1]
	}

	if matches := blackRe.FindStringSubmatch(pgn); len(matches) > 0 {
		g.Black = matches[1]
	}
}

// parse result
func (g *Game) parseResult(pgn string) {
	if matches := resultRe.FindStringSubmatch(pgn); len(matches) > 0 {
		switch matches[1] {
		case "1-0":
			g.Result = 1.0
		case "0-1":
			g.Result = 0.0
		case "1/2-1/2":
			g.Result = 0.5
		}
	}
}

// parse starting fen
func (g *Game) parseFEN(pgn string) {
	if matches := fenRe.FindStringSubmatch(pgn); len(matches) > 0 {
		g.FEN = matches[1]
	} else {
		g.FEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	}
}

// parse moves
func (g *Game) parseMoves(pgn string) {
	movetext := extractMovetext(pgn)

	pairs, err := tokenizeWithComments(movetext)
	if err != nil {
		log.Fatalf("failed to tokenize movetext: %v", err)
	}

	g.Moves = make([]string, len(pairs))

	b := g.Board.Preserve()

	for i, p := range pairs {
		m := ParseSAN(&b, p.san)
		if m == board.NOMOVE {
			log.Fatalf("move %d (%q): unparseable SAN", i+1, p.san)
		}

		g.Moves[i] = m.String()

		b.MakeMove(m, false)
		b.Ply++
	}
}

// returns everything in the PGN after the last header tag
func extractMovetext(pgn string) string {
	matches := taglineRe.FindAllStringIndex(pgn, -1)
	if len(matches) == 0 {
		return pgn
	}
	last := matches[len(matches)-1]
	return pgn[last[1]:]
}

// replays moves on board returns one line per ply
func (g *Game) GenerateLines() ([]string, error) {
	lines := make([]string, 0, len(g.Moves))

	for i, uci := range g.Moves {
		m := g.Board.ParseMove(uci)
		if m == board.NOMOVE {
			return nil, fmt.Errorf("move %d (%q): unparseable in position\n%s", i+1, uci, g.Board.FEN())
		}

		isTacticalMove := m.IsCapture() || m.Promoted(g.Board.Side) != board.Empty

		if ok := g.Board.MakeMove(m, false); !ok {
			return nil, fmt.Errorf("move %d (%q): illegal move\n%s", i+1, uci, g.Board.FEN())
		}
		g.Board.Ply++

		// skip positions immediately following a capture or promotion
		if isTacticalMove {
			continue
		}

		// check for king safety
		kingPiece := board.WK
		attackerColor := board.Black
		if g.Board.Side == board.Black {
			kingPiece = board.BK
			attackerColor = board.White
		}

		kingSq := board.Square(g.Board.PieceBitBoards[kingPiece].GetLSB())
		inCheck := g.Board.IsSquareAttacked(kingSq, attackerColor)
		if inCheck {
			continue
		}

		lines = append(lines, fmt.Sprintf("%s", g.Board.FEN()))
	}

	return lines, nil
}

// splits a file containing multiple concatenated PGN games into individual games
func SplitGames(data string) []string {
	idx := eventTagPattern.FindAllStringIndex(data, -1)
	if len(idx) == 0 {
		return nil
	}

	games := make([]string, 0, len(idx))
	for i, m := range idx {
		start := m[0]
		end := len(data)
		if i+1 < len(idx) {
			end = idx[i+1][0]
		}
		game := strings.TrimSpace(data[start:end])
		if game != "" {
			games = append(games, game)
		}
	}
	return games
}

// reads a file containing n concatenated PGN games
func ParseGamesFromFile(source string) ([]Game, []error) {
	data, err := os.ReadFile(source)
	if err != nil {
		return nil, []error{fmt.Errorf("reading file: %w", err)}
	}

	pgnGames := SplitGames(string(data))
	totalGames := len(pgnGames)

	var allGames []Game
	var errs []error
	for i, pgnStr := range pgnGames {
		game, err := NewGame(pgnStr)
		if err != nil {
			errs = append(errs, fmt.Errorf("game %d: %w", i+1, err))
			continue
		}

		allGames = append(allGames, game)
		fmt.Printf("\r\033[K[game %d/%d] parsed", i+1, totalGames)
	}

	fmt.Println()
	os.Stdout.Sync()

	return allGames, errs
}

func (g *Game) Print() {
	fmt.Printf("White: %s\t Black: %s\n", g.White, g.Black)

	switch g.Result {
	case 0:
		fmt.Printf("Result: Black Won\n")
	case 0.5:
		fmt.Printf("Result: Draw\n")
	default:
		fmt.Printf("Result: White Won\n")
	}

	fmt.Printf("FEN: %s\n", g.FEN)

	for _, move := range g.Moves {
		fmt.Printf("%s ", move)
	}

	fmt.Printf("\n")

}
