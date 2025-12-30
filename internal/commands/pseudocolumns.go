package commands

import "strings"

type pseudoColumn struct {
	ID   string
	Name string
	Kind string
}

var (
	pseudoColumnNotYet = pseudoColumn{ID: "not-yet", Name: "Not Yet", Kind: "triage"}
	pseudoColumnMaybe  = pseudoColumn{ID: "maybe", Name: "Maybe?", Kind: "not_now"}
	pseudoColumnDone   = pseudoColumn{ID: "done", Name: "Done", Kind: "closed"}
)

func pseudoColumnsInBoardOrder() []pseudoColumn {
	return []pseudoColumn{pseudoColumnNotYet, pseudoColumnMaybe, pseudoColumnDone}
}

func pseudoColumnObject(c pseudoColumn) map[string]interface{} {
	return map[string]interface{}{
		"id":     c.ID,
		"name":   c.Name,
		"kind":   c.Kind,
		"pseudo": true,
	}
}

func parsePseudoColumnID(id string) (pseudoColumn, bool) {
	switch strings.ToLower(strings.TrimSpace(id)) {
	case "not-yet", "not_yet", "notyet", "triage":
		return pseudoColumnNotYet, true
	case "maybe", "maybe?", "not-now", "not_now", "notnow":
		return pseudoColumnMaybe, true
	case "done", "closed", "close":
		return pseudoColumnDone, true
	default:
		return pseudoColumn{}, false
	}
}
