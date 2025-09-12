package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/MihaiZegheru/cfr/internal"
	"github.com/spf13/cobra"
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
		problems := make(map[string]internal.ProblemEntry)
		doc.Find("tr").Each(func(i int, s *goquery.Selection) {
			td := s.Find("td.id")
			if td.Length() == 0 {
				return
			}
			a := td.Find("a")
			probID := strings.TrimSpace(a.Text())
			href, exists := a.Attr("href")
			if probID != "" && exists && strings.HasPrefix(href, "/contest/"+id+"/problem/") {
				// Fetch sample tests for this problem
				probURL := "https://codeforces.com" + href
				tests := []internal.TestCase{}
				// Use the same client and headers as for the contest page
				probReq, err := http.NewRequest("GET", probURL, nil)
				if err == nil {
					probReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")
					probReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
					probReq.Header.Set("Accept-Language", "en-US,en;q=0.9")
					probReq.Header.Set("Referer", url)
					resp2, err := client.Do(probReq)
					if err == nil && resp2.StatusCode == 200 {
						defer resp2.Body.Close()
						doc2, err := goquery.NewDocumentFromReader(resp2.Body)
						if err == nil {
							var inputs, outputs []string
							doc2.Find("div.sample-test div.input pre").Each(func(i int, s *goquery.Selection) {
								input := ""
								s.Contents().Each(func(_ int, node *goquery.Selection) {
									if goquery.NodeName(node) == "br" {
										input += "\n"
									} else {
										input += node.Text()
									}
								})
								if input == "" {
									input = s.Text()
								}
								inputs = append(inputs, strings.TrimSpace(input))
							})
							doc2.Find("div.sample-test div.output pre").Each(func(i int, s *goquery.Selection) {
								output := ""
								s.Contents().Each(func(_ int, node *goquery.Selection) {
									if goquery.NodeName(node) == "br" {
										output += "\n"
									} else {
										output += node.Text()
									}
								})
								if output == "" {
									output = s.Text()
								}
								outputs = append(outputs, strings.TrimSpace(output))
							})
							for i := 0; i < len(inputs) && i < len(outputs); i++ {
								tests = append(tests, internal.TestCase{Input: inputs[i], Output: outputs[i]})
							}
						}
					}
				}
				problems[probID] = internal.ProblemEntry{URL: probURL, Tests: tests}
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
			fmt.Printf("%s: %s (%d tests)\n", k, v.URL, len(v.Tests))
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
			"c":      ".c",
			"cpp":    ".cpp",
			"c++":    ".cpp",
			"rust":   ".rs",
			"python": ".py",
			"py":     ".py",
			"go":     ".go",
			"java":   ".java",
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
