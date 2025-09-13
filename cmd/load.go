package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"regexp"
	"html"

	"github.com/PuerkitoBio/goquery"
	"github.com/MihaiZegheru/cfr/internal"
	"github.com/spf13/cobra"

)


import (
	// ...existing imports...
	htmlmd "github.com/JohannesKaufmann/html-to-markdown"
)

// htmlToMarkdown converts Codeforces problem HTML to readable markdown, including math.
func htmlToMarkdown(htmlStr string) string {
       // Pre-process: robustly replace <span class="tex-math">...</span> (even with nested tags/newlines) with $...$
       mathSpanRe := regexp.MustCompile(`(?s)<span class="tex-math">(.*?)</span>`)
       htmlStr = mathSpanRe.ReplaceAllStringFunc(htmlStr, func(s string) string {
	       m := mathSpanRe.FindStringSubmatch(s)
	       if len(m) == 2 {
		       inner := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(m[1], " ")
		       inner = regexp.MustCompile(`\s+`).ReplaceAllString(inner, " ")
		       inner = strings.TrimSpace(inner)
		       inner = strings.ReplaceAll(inner, "\\\\", "\\")
		       return "$" + inner + "$"
	       }
	       return s
       })
       // Pre-process: <sup class="tex-math">...</sup> and <sup>...</sup> to $^{...}$
       supMathRe := regexp.MustCompile(`(?s)<sup class="tex-math">(.*?)</sup>`)
       htmlStr = supMathRe.ReplaceAllStringFunc(htmlStr, func(s string) string {
	       m := supMathRe.FindStringSubmatch(s)
	       if len(m) == 2 {
		       inner := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(m[1], " ")
		       inner = regexp.MustCompile(`\s+`).ReplaceAllString(inner, " ")
		       inner = strings.TrimSpace(inner)
		       inner = strings.ReplaceAll(inner, "\\\\", "\\")
		       return "$^{" + inner + "}$"
	       }
	       return s
       })
       supPlainRe := regexp.MustCompile(`(?s)<sup>(.*?)</sup>`)
       htmlStr = supPlainRe.ReplaceAllStringFunc(htmlStr, func(s string) string {
	       m := supPlainRe.FindStringSubmatch(s)
	       if len(m) == 2 {
		       inner := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(m[1], " ")
		       inner = regexp.MustCompile(`\s+`).ReplaceAllString(inner, " ")
		       inner = strings.TrimSpace(inner)
		       return "$^{" + inner + "}$"
	       }
	       return s
       })
       // Pre-process: <i class="tex-math">...</i> to $...$
       iMathRe := regexp.MustCompile(`(?s)<i class="tex-math">(.*?)</i>`)
       htmlStr = iMathRe.ReplaceAllStringFunc(htmlStr, func(s string) string {
	       m := iMathRe.FindStringSubmatch(s)
	       if len(m) == 2 {
		       inner := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(m[1], " ")
		       inner = regexp.MustCompile(`\s+`).ReplaceAllString(inner, " ")
		       inner = strings.TrimSpace(inner)
		       inner = strings.ReplaceAll(inner, "\\\\", "\\")
		       return "$" + inner + "$"
	       }
	       return s
       })

       converter := htmlmd.NewConverter("", true, nil)
       md, err := converter.ConvertString(htmlStr)
       if err != nil {
	       return htmlStr // fallback to raw HTML if conversion fails
       }
       // Post-process: convert $$$...$$$ to $...$ for math mode
       md = strings.ReplaceAll(md, "$$$", "$")

       // Fix common Codeforces math HTML entities and symbols (but preserve LaTeX commands)
       mathReplacements := map[string]string{
	       "&le;": "≤", "&ge;": "≥", "&lt;": "<", "&gt;": ">", "&ne;": "≠", "&leq;": "≤", "&geq;": "≥",
       }
       for k, v := range mathReplacements {
	       md = strings.ReplaceAll(md, k, v)
       }

       // Section headings: Input, Output, Note, Examples, etc.
       sectionHeaders := []string{"Input", "Output", "Note", "Examples", "Example", "Interaction"}
       for _, sec := range sectionHeaders {
	       re := regexp.MustCompile(`(?m)^` + sec + `\n`)
	       md = re.ReplaceAllString(md, "## "+sec+"\n")
       }
       // Bold for time/memory limits
       limitHeaders := []string{"time limit per test", "memory limit per test", "input", "output"}
       for _, lim := range limitHeaders {
	       re := regexp.MustCompile(`(?mi)^(`+lim+`)(\n|:| )`)
	       md = re.ReplaceAllString(md, "**$1** ")
       }

       // Only fence clear sample/case blocks (blocks of numbers or code) after headings
       lines := strings.Split(md, "\n")
       var out []string
       inSample := false
       for i := 0; i < len(lines); i++ {
	       line := lines[i]
	       trimmed := strings.TrimSpace(line)
	       // Detect start of sample block: after heading, next non-empty, non-heading line that looks like data
	       if !inSample && i > 0 && (strings.HasPrefix(lines[i-1], "## Input") || strings.HasPrefix(lines[i-1], "## Output") || strings.HasPrefix(lines[i-1], "## Example")) && trimmed != "" && !strings.HasPrefix(trimmed, "## ") {
		       // Heuristic: if line contains mostly digits, symbols, or is indented, treat as sample
		       if regexp.MustCompile(`^[\d\s\-\+\*/\\.\(\)\[\]:;,A-Za-z]+$`).MatchString(trimmed) || strings.HasPrefix(line, "    ") {
			       out = append(out, "```text")
			       inSample = true
		       }
	       }
	       // End sample block at next heading or blank line
	       if inSample && (trimmed == "" || strings.HasPrefix(trimmed, "## ")) {
		       out = append(out, "```")
		       inSample = false
	       }
	       out = append(out, line)
       }
       if inSample {
	       out = append(out, "```")
       }

       md = strings.Join(out, "\n")
       // Remove excessive blank lines
       md = regexp.MustCompile(`\n{3,}`).ReplaceAllString(md, "\n\n")
       // Clean up trailing whitespace
       md = regexp.MustCompile(`[ \t]+\n`).ReplaceAllString(md, "\n")

       // Replace double backslashes with single backslash inside $...$ math blocks only
       md = regexp.MustCompile(`\$([^$]+)\$`).ReplaceAllStringFunc(md, func(s string) string {
	       // s is like "$...$"
	       content := s[1:len(s)-1] // remove the $ at start and end
	       content = strings.ReplaceAll(content, "\\\\", "\\")
	       // Convert \[ ... \] to \left[ ... \right] for KaTeX compatibility
	       content = regexp.MustCompile(`^\\\[(.*)\\\]$`).ReplaceAllString(content, `\\left[$1\\right]`)
	       return "$" + content + "$"
       })



	// Remove any lingering ${1} artifacts
	md = strings.ReplaceAll(md, "${1}", "")
	// Merge consecutive math blocks (e.g., $...$$...$ -> $... ...$)
	md = regexp.MustCompile(`\$\$`).ReplaceAllString(md, "")
	md = regexp.MustCompile(`\$\s+\$`).ReplaceAllString(md, " ")

       // Add a space after a math block if the next char is a letter/number (avoid $n$of)
       md = regexp.MustCompile(`\$([^
$]+)\$([a-zA-Z0-9])`).ReplaceAllString(md, "$${1}$ $2")

       // Final cleanup: remove any lingering ${1} artifacts
       md = strings.ReplaceAll(md, "${1}", "")
       // Final pass: ensure all double backslashes inside $...$ math blocks are single backslash
       md = regexp.MustCompile(`\$([^$]+)\$`).ReplaceAllStringFunc(md, func(s string) string {
	       content := s[1:len(s)-1]
	       content = strings.ReplaceAll(content, "\\\\", "\\")
	       return "$" + content + "$"
       })
	// Remove $x$ x or x $x$ (where x is a single variable or number)
	md = regexp.MustCompile(`\$([a-zA-Z0-9])\$ ?([a-zA-Z0-9])`).ReplaceAllString(md, "$1$2")
	md = regexp.MustCompile(`([a-zA-Z0-9]) ?\$([a-zA-Z0-9])\$`).ReplaceAllString(md, "$1$2")
	// Remove $x$ where x is a single variable/number and surrounded by spaces or punctuation
	md = regexp.MustCompile(`([\s\(\[]|^)\$([a-zA-Z0-9])\$([\s\)\]\.,;:]|$)`).ReplaceAllString(md, "$1$2$3")
	// Remove stray $ at the end or start of words/numbers not part of valid math blocks
	md = regexp.MustCompile(`([a-zA-Z0-9])\$([\s\.,;:\)\]\}]|$)`).ReplaceAllString(md, "$1$2")
	md = regexp.MustCompile(`(^|[\s\(\[\{])\$([a-zA-Z0-9])`).ReplaceAllString(md, "$1$2")
	return strings.TrimSpace(md) + "\n"
}
var loadCmd = &cobra.Command{
    Use:   "load <ID>",
    Short: "Load a problem by ID",
    Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfrDir := ".cfr"
		if _, err := os.Stat(cfrDir); os.IsNotExist(err) {
			fmt.Println("No .cfr directory found. Please run 'cfr init' first in this folder.")
			return
		}
		id := args[0]
		// Load current state if exists
		var prevState internal.ProblemsState
		prevState, _ = internal.LoadProblemsState()
		if prevState.ContestID != "" && prevState.ContestID != id {
			fmt.Printf("A different contest (%s) is already loaded. Please start a new workspace to load another contest.\n", prevState.ContestID)
			return
		}
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
	markdowns := make(map[string]string)
	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
            td := s.Find("td.id")
            if td.Length() == 0 {
                return
            }
            a := td.Find("a")
            probID := strings.TrimSpace(a.Text())
            href, exists := a.Attr("href")
			// Extract problem name from contest page row
			probName := ""
			s.Find("div[style*='float: left;'] a").EachWithBreak(func(_ int, nameA *goquery.Selection) bool {
				nameHref, _ := nameA.Attr("href")
				if nameHref == href {
					nameA.Contents().Each(func(_ int, n *goquery.Selection) {
						if goquery.NodeName(n) == "#text" {
							probName += n.Text()
						}
					})
					probName = strings.TrimSpace(probName)
					return false // break
				}
				return true
			})
				if probID != "" && exists && strings.HasPrefix(href, "/contest/"+id+"/problem/") {
					// Fetch sample tests for this problem
					probURL := "https://codeforces.com" + href
					tests := []internal.TestCase{}
					var problemMarkdown string
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
				       // Extract problem statement HTML and convert to Markdown
				       statementHtml, err := doc2.Find("div.problem-statement").Html()
				       if err == nil && statementHtml != "" {
					       problemMarkdown = htmlToMarkdown(statementHtml)
				       }
								// ...existing code for sample test extraction...
								var inputs, outputs []string
								doc2.Find("div.sample-test div.input pre").Each(func(i int, s *goquery.Selection) {
									// If there are <div>s, join their text with \n
									divs := s.Find("div")
									if divs.Length() > 0 {
										var lines []string
										divs.Each(func(_ int, div *goquery.Selection) {
											lines = append(lines, strings.TrimRight(div.Text(), "\r\n "))
										})
										inputs = append(inputs, strings.Join(lines, "\n"))
									} else {
										htmlStr, err := s.Html()
										if err != nil {
											inputs = append(inputs, strings.TrimSpace(s.Text()))
											return
										}
										// Replace <br> and <br/> with \n
										htmlStr = strings.ReplaceAll(htmlStr, "<br>", "\n")
										htmlStr = strings.ReplaceAll(htmlStr, "<br/>", "\n")
										htmlStr = strings.ReplaceAll(htmlStr, "<br />", "\n")
										// Remove all other tags
										re := regexp.MustCompile(`<[^>]+>`)
										htmlStr = re.ReplaceAllString(htmlStr, "")
										// Unescape HTML entities
										htmlStr = html.UnescapeString(htmlStr)
										inputs = append(inputs, strings.TrimSpace(htmlStr))
									}
								})
								doc2.Find("div.sample-test div.output pre").Each(func(i int, s *goquery.Selection) {
									htmlStr, err := s.Html()
									if err != nil {
										outputs = append(outputs, strings.TrimSpace(s.Text()))
										return
									}
									htmlStr = strings.ReplaceAll(htmlStr, "<br>", "\n")
									htmlStr = strings.ReplaceAll(htmlStr, "<br/>", "\n")
									htmlStr = strings.ReplaceAll(htmlStr, "<br />", "\n")
									re := regexp.MustCompile(`<[^>]+>`)
									htmlStr = re.ReplaceAllString(htmlStr, "")
									htmlStr = html.UnescapeString(htmlStr)
									outputs = append(outputs, strings.TrimSpace(htmlStr))
								})
								for i := 0; i < len(inputs) && i < len(outputs); i++ {
									tests = append(tests, internal.TestCase{Input: inputs[i], Output: outputs[i]})
								}
							}
						}
					}
					problems[probID] = internal.ProblemEntry{URL: probURL, Name: probName, Tests: tests}
					// Store markdown for writing after directory creation
					if probName != "" && problemMarkdown != "" {
						problems[probID] = internal.ProblemEntry{
							URL: probURL,
							Name: probName,
							Tests: tests,
							// Add a new field if needed for markdown, or handle after folder creation
						}
						// We'll write the markdown after all folders are created below
						// Use a map to store markdown if needed for all problems
						markdowns[probID] = problemMarkdown
					}

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
		fmt.Printf("Loaded %d problems for contest %s.\n", len(problems), id)

		// Only create files/folders if this is the first load (not a reload)
		if prevState.ContestID == "" {
			   configPath := ".cfr/config.json"
			   lang := "cpp"
			   ext := ".cpp"
			   if f, err := os.Open(configPath); err == nil {
				   defer f.Close()
				   var cfg struct {
					   DefaultLanguage string `json:"default_language"`
				   }
				   dec := json.NewDecoder(f)
				   if err := dec.Decode(&cfg); err == nil {
					   if cfg.DefaultLanguage != "" {
						   lang = strings.ToLower(cfg.DefaultLanguage)
					   }
				   }
			   }
			   extMap := map[string]string{
				   "c": ".c",
				   "cpp": ".cpp",
				   "c++": ".cpp",
				   "python": ".py",
				   "py": ".py",
				   "go": ".go",
			   }
			   if e, ok := extMap[lang]; ok {
				   ext = e
			   }
			   // Write markdowns after all problems are processed
			   for pid, prob := range problems {
				   dirName := fmt.Sprintf("%s. %s", pid, prob.Name)
				   os.MkdirAll(dirName, 0755)
				   inPath := dirName + string(os.PathSeparator) + "in.txt"
				   outPath := dirName + string(os.PathSeparator) + "out.txt"
				   os.WriteFile(inPath, []byte{}, 0644)
				   os.WriteFile(outPath, []byte{}, 0644)
				   srcPath := dirName + string(os.PathSeparator) + "main" + ext
				   if _, err := os.Stat(srcPath); os.IsNotExist(err) {
					   f, err := os.Create(srcPath)
					   if err == nil {
						   f.Close()
					   }
				   }
				  // Write markdown if available
				  if markdowns != nil {
					  if md, ok := markdowns[pid]; ok && md != "" {
						  mdPath := dirName + string(os.PathSeparator) + "task.md"
						  os.WriteFile(mdPath, []byte(md), 0644)
					  }
				  }
			   }
			   fmt.Printf("Created source and IO files for problems.\n")
		} else {
			fmt.Println("Only sample tests were updated. No files or folders were changed.")
		}
		fmt.Println("Done.")
	},
}

func init() {
	rootCmd.AddCommand(loadCmd)
}
