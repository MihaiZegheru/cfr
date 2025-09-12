package cmd

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"os"
	"strings"
	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
	"github.com/MihaiZegheru/cfr/internal"
)

var customTest bool

var testCmd = &cobra.Command{
	Use:   "test <problem_ID>",
	Short: "Test a problem by ID",
	Args:  cobra.ExactArgs(1),
       Run: func(cmd *cobra.Command, args []string) {
	       state, err := internal.LoadProblemsState()
	       if err != nil || state.ContestID == "" {
		       fmt.Println("No contest ID loaded. Please run 'cfr load <ID>' first.")
		       return
	       }
	       contestID := state.ContestID
	       problemID := args[0]
	       url := fmt.Sprintf("https://codeforces.com/contest/%s/problem/%s", contestID, problemID)
	       fmt.Printf("Fetching problem from: %s\n", url)
			client := &http.Client{}
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				fmt.Printf("Failed to create request: %v\n", err)
				return
			}
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("Failed to fetch problem: %v\n", err)
				return
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Failed to read response: %v\n", err)
				return
			}

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
			if err != nil {
				fmt.Printf("Failed to parse HTML: %v\n", err)
				return
			}

			var inputs []string
			var outputs []string
			doc.Find("div.sample-test div.input pre").Each(func(i int, s *goquery.Selection) {
				inputs = append(inputs, s.Text())
			})
			doc.Find("div.sample-test div.output pre").Each(func(i int, s *goquery.Selection) {
				outputs = append(outputs, s.Text())
			})

			if len(inputs) == 0 || len(outputs) == 0 {
				fmt.Println("Failed to find input or output examples on the page.")
				return
			}

			inputText := strings.Join(inputs, "\n\n")
			outputText := strings.Join(outputs, "\n\n")

			err = os.WriteFile("in.txt", []byte(inputText), 0644)
			if err != nil {
				fmt.Printf("Failed to write in.txt: %v\n", err)
				return
			}
			err = os.WriteFile("out.txt", []byte(outputText), 0644)
			if err != nil {
				fmt.Printf("Failed to write out.txt: %v\n", err)
				return
			}

			fmt.Printf("Wrote %d input(s) to in.txt and %d output(s) to out.txt.\n", len(inputs), len(outputs))
		},
}

func init() {
	testCmd.Flags().BoolVarP(&customTest, "custom", "c", false, "Run against custom test")
	rootCmd.AddCommand(testCmd)
}
