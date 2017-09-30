package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "zgen",
	Short: "zgen is a helper tool for zen web framework",
}

func main() {
	rootCmd.Execute()
}
