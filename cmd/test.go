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
	       if len(prob.Tests) == 0 {
		       fmt.Printf("No sample tests found for problem %s.\n", problemID)
		       return
	       }
	       // Detect language from config
	       lang := "cpp"
	       configPath := ".cfr/config.json"
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
		       fmt.Println("No valid language set in .cfr/config.json. Cannot test.")
		       return
	       }
	       sourceFile := problemID + ext
	       // Compile the solution
	       var execCmd string
	       var execArgs []string
	       var binName string
	       switch lang {
	       case "cpp", "c++":
		       binName = problemID + ".exe"
		       execCmd = "g++"
		       execArgs = []string{"-O2", "-std=c++17", sourceFile, "-o", binName}
	       case "c":
		       binName = problemID + ".exe"
		       execCmd = "gcc"
		       execArgs = []string{"-O2", sourceFile, "-o", binName}
	       case "rust":
		       binName = problemID + ".exe"
		       execCmd = "rustc"
		       execArgs = []string{sourceFile, "-o", binName}
	       case "go":
		       binName = problemID + ".exe"
		       execCmd = "go"
		       execArgs = []string{"build", "-o", binName, sourceFile}
	       case "java":
		       binName = problemID + ".class"
		       execCmd = "javac"
		       execArgs = []string{sourceFile}
	       case "python", "py":
		       binName = ""
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
		       runCmd = "python"
		       runArgs = []string{sourceFile}
	       } else if lang == "java" {
		       runCmd = "java"
		       runArgs = []string{"-cp", ".", strings.TrimSuffix(sourceFile, ".java")}
	       } else {
		       runCmd = "./" + binName
		       runArgs = []string{}
	       }
	       // Run each test case
	       fmt.Printf("Running %d sample test(s)...\n", len(prob.Tests))
	       for i, tc := range prob.Tests {
		       // Write input to temp file
		       inFile := fmt.Sprintf(".cfr/tmp_input_%d.txt", i)
		       os.WriteFile(inFile, []byte(tc.Input), 0644)
		       output, err := runWithInput(runCmd, runArgs, inFile)
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
