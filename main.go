package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/celestiaorg/celestia-node/nodebuilder"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
)

var (
	dir              = ".nowreact"
	defaultNamespace = "nowreact"

	// Embed the static directory
	//go:embed static
	staticFS embed.FS
)

const (
	size          = 16
	stateFileName = "state.json"
)

func init() {
	if envDir := os.Getenv("NOWREACT_DIR"); envDir != "" {
		dir = envDir
	}
}

type State struct {
	Height    int      `json:"height"`
	Namespace string   `json:"namespace"`
	Square    []string `json:"emojis"`
}

func DefaultState() *State {
	return &State{
		Height:    1,
		Namespace: defaultNamespace,
		Square:    make([]string, size*size),
	}
}

type EmojiSubmission struct {
	Emoji    string `json:"emoji"`
	Index int    `json:"location"`
}

func LoadState() (*State, error) {
	filename := filepath.Join(dir, stateFileName)
	// Check if the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// File doesn't exist, return the default state
		return DefaultState(), nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var state *State
	if err := json.Unmarshal(data, state); err != nil {
		return nil, err
	}
	return state, nil
}

func SaveState(state State) error {
	filename := filepath.Join(dir, stateFileName)
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling state: %v", err)
	}

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing state file: %v", err)
	}

	return nil
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]int{"height": square.Height})
}

func getSquareHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(square.Emojis)
}

func postEmojiHandler(w http.ResponseWriter, r *http.Request) {
	var submission EmojiSubmission
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if submission.Index < 0 || submission.Index >= len(square.Emojis) {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	square.Emojis[submission.Index] = submission.Emoji
	w.WriteHeader(http.StatusOK)
}

func main() {
	state, err := LoadState()
	if err != nil {
		log.Fatalf("error loading state: %v", err)
	}

	cfg := nodebuilder.DefaultConfig(node.Light)
	nodebuilder.Init(*cfg, dir, node.Light)

	// Create a file server for the embedded static files
	staticHandler := http.FileServer(http.FS(staticFS))

	// Serve static files from the root path
	http.Handle("/", http.StripPrefix("/", staticHandler))

	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/getSquare", getSquareHandler)
	http.HandleFunc("/postEmoji", postEmojiHandler)

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
