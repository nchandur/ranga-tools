package tournament

import "github.com/nchandur/ranga-tools/internal/tournament"

var selfTestCmd = newTournamentCmd(
	"selftest",
	"play registered engine against itself",
	tournament.RunSelfTest,
)
