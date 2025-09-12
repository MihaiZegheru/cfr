package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
)


	const cfrDir = ".cfr"
	const stateFile = "problems.json"


func getStatePath() string {
	if _, err := os.Stat(cfrDir); err == nil {
		return filepath.Join(cfrDir, stateFile)
	}
	return stateFile
}

type TestCase struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type ProblemEntry struct {
	URL   string     `json:"url"`
	Name  string     `json:"name"`
	Tests []TestCase `json:"tests"`
}

type ProblemsState struct {
	ContestID string                  `json:"contest_id"`
	Problems  map[string]ProblemEntry `json:"problems"`
}

func SaveProblemsState(state ProblemsState) error {
	f, err := os.Create(getStatePath())
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	return err
}

func LoadProblemsState() (ProblemsState, error) {
	var state ProblemsState
	data, err := os.ReadFile(getStatePath())
	if err != nil {
		return state, err
	}
	err = json.Unmarshal(data, &state)
	return state, err
}
