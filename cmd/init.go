package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize CFR state in the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		cfrDir := ".cfr"
		if _, err := os.Stat(cfrDir); err == nil {
			fmt.Println(".cfr directory already exists. Initialization aborted.")
			return
		}
		err := os.MkdirAll(cfrDir, 0755)
		if err != nil {
			fmt.Printf("Failed to create .cfr directory: %v\n", err)
			return
		}
		   // Create default config.json if not exists
		   configPath := cfrDir + string(os.PathSeparator) + "config.json"
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			f, err := os.Create(configPath)
			if err != nil {
				fmt.Printf("Failed to create %s: %v\n", configPath, err)
			} else {
				f.WriteString(`{
  "default_language": "cpp",
  "languages": {},
  "executables": {
	"cpp": "g++",
	"c": "gcc",
	"go": "go",
	"python": "python"
  }
}
`)
				f.Close()
				fmt.Printf("Created default config at %s\n", configPath)
			}
		}
		fmt.Println("Initialized CFR state in .cfr directory.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
