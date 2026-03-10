package commands

import "testing"

func FuzzParsePseudoColumnID(f *testing.F) {
	f.Add("")
	f.Add("not-now")
	f.Add("NOT_NOW")
	f.Add("maybe")
	f.Add("maybe?")
	f.Add("triage")
	f.Add("done")
	f.Add("closed")
	f.Add("close")
	f.Add("random-string")
	f.Add("  done  ")

	f.Fuzz(func(t *testing.T, id string) {
		parsePseudoColumnID(id) // must not panic
	})
}

func FuzzNormalizeSkillPath(f *testing.F) {
	f.Add("")
	f.Add("~/skills")
	f.Add("~/skills/fizzy")
	f.Add("~/skills/fizzy/SKILL.md")
	f.Add("/tmp/custom/path")
	f.Add("/tmp/custom/path/fizzy")
	f.Add("relative/path")
	f.Add("file.md")

	f.Fuzz(func(t *testing.T, path string) {
		normalizeSkillPath(path) // must not panic
	})
}

func FuzzExpandPath(f *testing.F) {
	f.Add("")
	f.Add("~")
	f.Add("~/")
	f.Add("~/foo/bar")
	f.Add("/absolute/path")
	f.Add("relative/path")

	f.Fuzz(func(t *testing.T, path string) {
		expandPath(path) // must not panic
	})
}
