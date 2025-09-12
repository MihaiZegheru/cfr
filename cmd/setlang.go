package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

import "github.com/spf13/cobra"

var setLangCmd = &cobra.Command{
	Use:   "set-lang <PROBLEM_ID> <language>",
	Short: "Set the language for a specific problem and create the source file if needed",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		problemID := args[0]
		lang := strings.ToLower(args[1])
		supported := map[string]string{
			"c": ".c",
			"cpp": ".cpp",
			"c++": ".cpp",
			"rust": ".rs",
			"python": ".py",
			"py": ".py",
			"go": ".go",
			"java": ".java",
		}
		ext, ok := supported[lang]
		if !ok {
			fmt.Printf("Unsupported language: %s\n", lang)
			return
		}
		// Load config
		configPath := ".cfr/config.json"
		type LangConfig struct {
			DefaultLanguage string            `json:"default_language"`
			Languages       map[string]string `json:"languages"`
			Executables     map[string]string `json:"executables"`
		}
		var cfg LangConfig
		cfg.Languages = map[string]string{}
		if f, err := os.Open(configPath); err == nil {
			defer f.Close()
			dec := json.NewDecoder(f)
			_ = dec.Decode(&cfg)
		}
		if cfg.Languages == nil {
			cfg.Languages = map[string]string{}
		}
		   // Check if already set to this language
		   if prev, ok := cfg.Languages[problemID]; ok && prev == lang {
			   fmt.Printf("[CFR] Language for problem %s is already set to '%s'. No changes made.\n", problemID, lang)
			   return
		   }
		   cfg.Languages[problemID] = lang
		   // Save config
		   f, err := os.Create(configPath)
		   if err != nil {
			   fmt.Printf("Failed to update config: %v\n", err)
			   return
		   }
		   enc := json.NewEncoder(f)
		   enc.SetIndent("", "  ")
		   if err := enc.Encode(cfg); err != nil {
			   fmt.Printf("Failed to write config: %v\n", err)
			   f.Close()
			   return
		   }
		   f.Close()
		   fmt.Printf("[CFR] Language for problem %s set to '%s'.\n", problemID, lang)
		// Create source file if needed
		// Find problem directory
		problemsStatePath := ".cfr/problems.json"
		type ProblemEntry struct {
			Name string `json:"name"`
		}
		type ProblemsState struct {
			Problems map[string]ProblemEntry `json:"problems"`
		}
		var state ProblemsState
		if data, err := os.ReadFile(problemsStatePath); err == nil {
			_ = json.Unmarshal(data, &state)
			if entry, ok := state.Problems[problemID]; ok {
				dirName := fmt.Sprintf("%s. %s", problemID, entry.Name)
				// Move non-empty old main.<ext> file to versions/, remove empty ones, except for the new one
				supportedExts := []string{".c", ".cpp", ".rs", ".py", ".go", ".java"}
				versionsDir := dirName + string(os.PathSeparator) + "versions"
				if err := os.MkdirAll(versionsDir, 0755); err != nil {
					fmt.Printf("[CFR] Warning: Could not create versions directory: %v\n", err)
				}
				for _, oldExt := range supportedExts {
					oldPath := dirName + string(os.PathSeparator) + "main" + oldExt
					versionPath := versionsDir + string(os.PathSeparator) + "main" + oldExt
					if oldExt != ext {
						if fi, err := os.Stat(oldPath); err == nil {
							if fi.Size() == 0 {
								if err := os.Remove(oldPath); err == nil {
									fmt.Printf("[CFR] Removed empty file: %s\n", oldPath)
								} else {
									fmt.Printf("[CFR] Warning: Could not remove %s: %v\n", oldPath, err)
								}
							} else {
								// Move non-empty file to versions/
								if err := os.Rename(oldPath, versionPath); err == nil {
									fmt.Printf("[CFR] Saved previous version: %s → %s\n", oldPath, versionPath)
								} else {
									fmt.Printf("[CFR] Warning: Could not move %s: %v\n", oldPath, err)
								}
							}
						}
					}
				}
				srcPath := dirName + string(os.PathSeparator) + "main" + ext
				versionRestore := versionsDir + string(os.PathSeparator) + "main" + ext
				if _, err := os.Stat(srcPath); os.IsNotExist(err) {
					// Try to restore from versions/
					if _, err := os.Stat(versionRestore); err == nil {
						if err := os.Rename(versionRestore, srcPath); err == nil {
							fmt.Printf("[CFR] Restored previous version for %s from versions/\n", srcPath)
						} else {
							fmt.Printf("[CFR] Warning: Could not restore %s: %v\n", srcPath, err)
						}
					} else {
						f, err := os.Create(srcPath)
						if err == nil {
							f.Close()
							fmt.Printf("[CFR] Created new source file: %s\n", srcPath)
						} else {
							fmt.Printf("[CFR] Error: Could not create %s: %v\n", srcPath, err)
						}
					}
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(setLangCmd)
}
