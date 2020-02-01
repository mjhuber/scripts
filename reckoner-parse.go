/*
This tool returns version info about releases in a reckoner course.
*/
package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)


// Course is a reckoner course file
type Course struct {
	Namespace *string          `yaml:"namespace,omitempty"`
	Context   *string          `yaml:"context,omitempty"`
	Charts    map[string]Chart `yaml:"charts"`
}

// Chart is a chart defined in a reckoner file
type Chart struct {
	Version   string  `yaml:"version"`
	Namespace *string `yaml:"namespace,omitempty"`
}

var (
	path     string
	releases []string
	all      bool
	rootCmd  = &cobra.Command{
		Use:   "rchk",
		Short: "rchk - get reckoner course info",
		Long:  "rchk - get reckoner course info",
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}
)

func main() {
	rootCmd.PersistentFlags().StringVarP(&path, "file", "f", "", "Path to reckoner course")
	rootCmd.PersistentFlags().StringSliceVarP(&releases, "release", "r", []string{}, "Names of releases to pull")
	rootCmd.PersistentFlags().BoolVarP(&all, "all", "a", false, "When true retrurns all releases")
	rootCmd.MarkPersistentFlagRequired("file")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() {
	course := Course{}
	file, err := ioutil.ReadFile(path)

	if err != nil {
		fmt.Printf("Error reading file: %v", err)
		os.Exit(1)
	}

	err = yaml.Unmarshal(file, &course)
	if err != nil {
		fmt.Printf("Error unmarshalling yaml: %v\n", err)
		os.Exit(1)
	}

	for name, chart := range course.Charts {
		if all || Contains(releases, name) {
			fmt.Printf("%s: %s\n", name, chart.Version)
		}
	}
}

// Contains returns true if name is in items
func Contains(items []string, name string) bool {
	for _, item := range items {
		if name == item {
			return true
		}
	}
	return false
}
