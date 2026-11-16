package tournament

import "github.com/nchandur/ranga-tools/internal/tournament"

var sprtCmd = newTournamentCmd(
	"sprt",
	"run SPRT between an RC engine and a base engine",
	tournament.RunSPRT,
)
