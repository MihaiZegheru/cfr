package cmd

import (
	"fmt"
	"os"
	"strings"
	"encoding/json"
	"github.com/spf13/cobra"
	"github.com/MihaiZegheru/cfr/internal"
)

var customTest bool

var testCmd = &cobra.Command{
	 Use:   "test <problem_ID>",
	 Short: "Test a problem by ID",
			 Long: `Test a problem by ID.

		By default, runs all sample tests for the problem.

		Use -c to run a custom test: input is read from in.txt and output is written to out.txt in the problem directory.

		Language selection:
			- The language for each problem can be set in .cfr/config.json:
				{
					"default_language": "cpp",
					"languages": {
						"A": "cpp",
						"B": "python"
					}
				}
			- Supported languages: cpp, c, rust, go, python, java
			- If a problem is not listed in 'languages', 'default_language' is used.
		`,
	 Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
		   state, err := internal.LoadProblemsState()
		   if err != nil || state.ContestID == "" {
			   fmt.Println("No contest ID loaded. Please run 'cfr load <ID>' first.")
			   return
		   }
		   problemID := args[0]
		   prob, ok := state.Problems[problemID]
		   if !ok {
			   fmt.Printf("Problem %s not found in state. Please run 'cfr load <ID>' again.\n", problemID)
			   return
		   }
		  // Detect language from config
		  lang := "cpp"
		  configPath := ".cfr/config.json"
		  type LangConfig struct {
			  DefaultLanguage string            `json:"default_language"`
			  Languages       map[string]string `json:"languages"`
			  Executables     map[string]string `json:"executables"`
		  }
		  var cfg LangConfig
		  if f, err := os.Open(configPath); err == nil {
			  defer f.Close()
			  dec := json.NewDecoder(f)
			  if err := dec.Decode(&cfg); err == nil {
				  if l, ok := cfg.Languages[problemID]; ok {
					  lang = strings.ToLower(l)
				  } else if cfg.DefaultLanguage != "" {
					  lang = strings.ToLower(cfg.DefaultLanguage)
				  }
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
			  fmt.Println("No valid language set in .cfr/config.json. Cannot test.")
			  return
		  }
		   // Find the problem directory (should match the format <problemID>. <name>)
		   probDir := ""
		   for id, entry := range state.Problems {
			   if id == problemID {
				   probDir = fmt.Sprintf("%s. %s", id, entry.Name)
				   break
			   }
		   }
		   if probDir == "" {
			   fmt.Printf("Directory for problem %s not found.\n", problemID)
			   return
		   }
		   sourceFile := probDir + string(os.PathSeparator) + "main" + ext
		   // Compile the solution
		   var execCmd string
		   var execArgs []string
		   var binName string
		   var binPath string
		  // Select executable from config or use default
		  defaultExecs := map[string]string{
			  "cpp": "g++",
			  "c": "gcc",
			  "go": "go",
			  "python": "python",
			  "py": "python",
		  }
		  getExec := func(lang string) string {
			  if cfg.Executables != nil {
				  if exe, ok := cfg.Executables[lang]; ok && exe != "" {
					  return exe
				  }
			  }
			  if exe, ok := defaultExecs[lang]; ok {
				  return exe
			  }
			  return ""
		  }
		  switch lang {
		  case "cpp", "c++":
			  binName = problemID + ".exe"
			  binPath = probDir + string(os.PathSeparator) + binName
			  execCmd = getExec("cpp")
			  execArgs = []string{"-O2", "-std=c++17", sourceFile, "-o", binPath}
		  case "c":
			  binName = problemID + ".exe"
			  binPath = probDir + string(os.PathSeparator) + binName
			  execCmd = getExec("c")
			  execArgs = []string{"-O2", sourceFile, "-o", binPath}
		  case "go":
			  binName = problemID + ".exe"
			  binPath = probDir + string(os.PathSeparator) + binName
			  execCmd = getExec("go")
			  execArgs = []string{"build", "-o", binPath, sourceFile}
		  case "python", "py":
			  binName = ""
			  binPath = ""
		  default:
			  fmt.Println("Language not supported for testing.")
			  return
		  }
		   if lang != "python" && lang != "py" {
			   fmt.Printf("Compiling %s...\n", sourceFile)
			   out, err := runAndCapture(execCmd, execArgs...)
			   if err != nil {
				   fmt.Printf("Compilation failed: %v\n%s\n", err, out)
				   return
			   }
			   fmt.Println("Compilation successful.")
		   }
		   // Prepare run command
		  var runCmd string
		  var runArgs []string
		  if lang == "python" || lang == "py" {
			  runCmd = getExec(lang)
			  runArgs = []string{sourceFile}
		  } else {
			  runCmd = "." + string(os.PathSeparator) + binName
			  runArgs = []string{}
		  }

		   if customTest {
			   // Use in.txt as input, write output to out.txt in the problem directory
			   inPath := probDir + string(os.PathSeparator) + "in.txt"
			   outPath := probDir + string(os.PathSeparator) + "out.txt"
			   if _, err := os.Stat(inPath); os.IsNotExist(err) {
				   fmt.Printf("%s not found. Please run 'cfr load <ID>' first.\n", inPath)
				   return
			   }
			   runCwd := ""
			   if lang != "python" && lang != "py" && lang != "java" {
				   runCwd = probDir
			   }
			   output, err := runWithInputCwd(runCmd, runArgs, inPath, runCwd)
			   if err != nil {
				   fmt.Printf("Execution failed: %v\n", err)
				   return
			   }
			   os.WriteFile(outPath, []byte(output), 0644)
			   fmt.Printf("Custom test complete. Output written to %s\n", outPath)
			   return
		   }

		   if len(prob.Tests) == 0 {
			   fmt.Printf("No sample tests found for problem %s.\n", problemID)
			   return
		   }
		   // Run each test case
		   fmt.Printf("Running %d sample test(s)...\n", len(prob.Tests))
		   for i, tc := range prob.Tests {
			   // Write input to temp file in the problem directory
			   inFile := probDir + string(os.PathSeparator) + fmt.Sprintf("tmp_input_%d.txt", i)
			   os.WriteFile(inFile, []byte(tc.Input), 0644)
			   // If running a compiled language, run from the problem directory
			   runCwd := ""
			   if lang != "python" && lang != "py" && lang != "java" {
				   runCwd = probDir
			   }
			   output, err := runWithInputCwd(runCmd, runArgs, inFile, runCwd)
			   // Clean up input file
			   os.Remove(inFile)
			   fmt.Printf("Test #%d:\n", i+1)
			   if err != nil {
				   fmt.Printf("  Execution failed: %v\n", err)
				   continue
			   }
			   // Normalize output for comparison: trim trailing spaces per line and ignore extra blank lines at end
			   normalize := func(s string) string {
				   lines := strings.Split(s, "\n")
				   var cleaned []string
				   for _, line := range lines {
					   cleaned = append(cleaned, strings.TrimRight(line, " \t\r"))
				   }
				   // Remove trailing empty lines
				   for len(cleaned) > 0 && cleaned[len(cleaned)-1] == "" {
					   cleaned = cleaned[:len(cleaned)-1]
				   }
				   return strings.Join(cleaned, "\n")
			   }
			   userOut := normalize(output)
			   expected := normalize(tc.Output)
			   if userOut == expected {
				   fmt.Println("  OK")
			   } else {
				   fmt.Println("  Wrong Answer")
				   fmt.Println("  Your output:")
				   fmt.Println(userOut)
				   fmt.Println("  Expected output:")
				   fmt.Println(expected)
			   }
		   }
       },
}

func init() {
	testCmd.Flags().BoolVarP(&customTest, "custom", "c", false, "Run against custom test")
	rootCmd.AddCommand(testCmd)
}
