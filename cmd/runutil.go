package cmd

import (
       "bytes"
       "os"
       "os/exec"
)

// runAndCapture runs a command and returns its combined output and error
func runAndCapture(cmd string, args ...string) (string, error) {
       c := exec.Command(cmd, args...)
       var out bytes.Buffer
       c.Stdout = &out
       c.Stderr = &out
       err := c.Run()
       return out.String(), err
}

// runWithInput runs a command, feeds it the contents of inputFile as stdin, and returns its stdout
func runWithInput(cmd string, args []string, inputFile string) (string, error) {
       c := exec.Command(cmd, args...)
       in, err := os.Open(inputFile)
       if err != nil {
               return "", err
       }
       defer in.Close()
       c.Stdin = in
       var out bytes.Buffer
       c.Stdout = &out
       c.Stderr = &out
       err = c.Run()
       return out.String(), err
}

// runWithInputCwd runs a command, feeds it the contents of inputFile as stdin, sets the working directory, and returns its stdout
func runWithInputCwd(cmd string, args []string, inputFile string, cwd string) (string, error) {
       c := exec.Command(cmd, args...)
       if cwd != "" {
               c.Dir = cwd
       }
       in, err := os.Open(inputFile)
       if err != nil {
               return "", err
       }
       defer in.Close()
       c.Stdin = in
       var out bytes.Buffer
       c.Stdout = &out
       c.Stderr = &out
       err = c.Run()
       return out.String(), err
}