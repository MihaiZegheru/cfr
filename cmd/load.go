package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
	"github.com/MihaiZegheru/cfr/internal"
)

var loadCmd = &cobra.Command{
	Use:   "load <ID>",
	Short: "Load a problem by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
				// Check for .cfr directory
				if _, err := os.Stat(".cfr"); os.IsNotExist(err) {
					fmt.Println("No .cfr directory found. Please run 'cfr init' first in this folder.")
					return
				}
				id := args[0]
				fmt.Printf("Loaded contest ID: %s\n", id)

				// Fetch and save problem IDs
				url := fmt.Sprintf("https://codeforces.com/contest/%s", id)
				client := &http.Client{}
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					fmt.Printf("Failed to create request for problems: %v\n", err)
					return
				}
				req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")
				resp, err := client.Do(req)
				if err != nil {
					fmt.Printf("Failed to fetch contest page: %v\n", err)
					return
				}
				defer resp.Body.Close()
				doc, err := goquery.NewDocumentFromReader(resp.Body)
				if err != nil {
					fmt.Printf("Failed to parse contest HTML: %v\n", err)
					return
				}
				problems := make(map[string]string)
				doc.Find("tr").Each(func(i int, s *goquery.Selection) {
					td := s.Find("td.id")
					if td.Length() == 0 {
						return
					}
					a := td.Find("a")
					probID := strings.TrimSpace(a.Text())
					href, exists := a.Attr("href")
					if probID != "" && exists && strings.HasPrefix(href, "/contest/"+id+"/problem/") {
						problems[probID] = href
					}
				})
				if len(problems) == 0 {
					fmt.Println("No problems found for this contest.")
					return
				}
				// Save unified state
				state := internal.ProblemsState{
					ContestID: id,
					Problems:  problems,
				}
				err = internal.SaveProblemsState(state)
				if err != nil {
					fmt.Printf("Failed to save problems state: %v\n", err)
					return
				}
				fmt.Println("Saved problems:")
				for k, v := range problems {
					fmt.Printf("%s: %s\n", k, v)
				}

				// Read config for language
				configPath := ".cfr/config.json"
				lang := ""
				if f, err := os.Open(configPath); err == nil {
					defer f.Close()
					var cfg struct{ Language string `json:"language"` }
					dec := json.NewDecoder(f)
					if err := dec.Decode(&cfg); err == nil {
						lang = strings.ToLower(cfg.Language)
					}
				}
				ext := map[string]string{
					"c": ".c",
					"cpp": ".cpp",
					"c++": ".cpp",
					"rust": ".rs",
					"python": ".py",
					"py": ".py",
					"go": ".go",
					"java": ".java",
				}[lang]
				if ext == "" {
					fmt.Println("No valid language set in .cfr/config.json. Skipping file creation.")
				} else {
					for id := range problems {
						fname := id + ext
						if _, err := os.Stat(fname); os.IsNotExist(err) {
							f, err := os.Create(fname)
							if err != nil {
								fmt.Printf("Failed to create %s: %v\n", fname, err)
							} else {
								f.Close()
								fmt.Printf("Created %s\n", fname)
							}
						}
					}
				}
			},
}

func init() {
	rootCmd.AddCommand(loadCmd)
}
