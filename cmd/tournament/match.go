package tournament

import "github.com/nchandur/ranga-tools/internal/tournament"

var matchCmd = newTournamentCmd(
	"match",
	"play a match between two engines",
	tournament.RunMatch,
)
