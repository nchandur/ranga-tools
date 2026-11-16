package tournament

import "github.com/nchandur/ranga-tools/internal/tournament"

var sprtCmd = newTournamentCmd(
	"sprt",
	"Run an SPRT between an RC engine and a base engine",
	tournament.RunSPRT,
)
