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
	"path/filepath"
	"strings"
)
//go run reckoner-parse.go -i kiva -r rbac-manager

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
	inventories []string
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
	rootCmd.PersistentFlags().StringSliceVarP(&inventories, "inventory", "i", []string{}, "Names of inventories to search")
	rootCmd.PersistentFlags().StringSliceVarP(&releases, "release", "r", []string{}, "Names of releases to pull")
	rootCmd.PersistentFlags().BoolVarP(&all, "all", "a", false, "When true retrurns all releases")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() {
	dir := os.Getenv("CUDDLEFISH_PROJECTS_DIR")
	err := filepath.Walk(dir,
					   func(path string, info os.FileInfo, err error) error {
						   if err != nil {
							   return err
						   }
						   //fmt.Println(path, info.Size(), info.Name())
						   if info.Name() == "course.yml" && (Matches(inventories,path)) {
							   //fmt.Println(path)
							   err := parseFile(path)
							   if err != nil {
								   fmt.Printf("\nError processing %s: %v\n\n",err)
							   }
						   }
						   return nil
					   })
	if err != nil {
		fmt.Println(err)
	}
}

func parseFile(path string) error {
	fmt.Println(path)
	course := Course{}
	file, err := ioutil.ReadFile(path)

	if err != nil {
		fmt.Printf("Error reading file: %v", err)
		return err
	}

	err = yaml.Unmarshal(file, &course)
	if err != nil {
		return err
	}

	for name, chart := range course.Charts {
		if all || Contains(releases, name) {
			fmt.Printf("\t%s: %s\n", name, chart.Version)
		}
	}
	return nil
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

// Matches returns true if any substrings in items are in name
func Matches(items []string, name string) bool {
	for _, item := range items {
		if item == "*" || strings.Contains(name, item) {
			return true
		}
	}
	return false
}
