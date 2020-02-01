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
	"net/http"
	goversion "github.com/mcuadros/go-version"
	"github.com/fatih/color"
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

// FairwindsStandards contains the fairwinds standards
type FairwindsStandards struct {
	Helm HelmStandards `yaml:"helm"`
}

// HelmStandards contain the helm standards from the docs-engineering repo
type HelmStandards struct {
	Reckoner         string              `yaml:"reckoner_version"`
	Helm             string              `yaml:"helm_version"`
	Repositories     map[string][]string `yaml:"repository"`
	Charts           map[string]Version  `yaml:"chart"`
	DeprecatedCharts []string            `yaml:"deprecated"`
}

// Version is a stringed version in semver format
type Version struct {
	Version string `yaml:"version"`
}

var (
	inventories  []string
	releases     []string
	all          bool
	standardsURL string
	rootCmd      = &cobra.Command{
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
	rootCmd.PersistentFlags().StringVarP(&standardsURL, "standards", "s", "https://a070b7d3-44ba-448d-ac3d-2b3237cc4468.s3.us-east-2.amazonaws.com/standards.yml", "URL for the fairwinds standards.")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() {
	dir := os.Getenv("CUDDLEFISH_PROJECTS_DIR")

	standards, err := getStandards()
	if err != nil {
		fmt.Printf("Error getting standards: %v\n")
		os.Exit(1)
	}

	err = filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			//fmt.Println(path, info.Size(), info.Name())
			if info.Name() == "course.yml" && (Matches(inventories, path)) {
				//fmt.Println(path)
				err := parseFile(path,standards)
				if err != nil {
					fmt.Printf("\nError processing %s: %v\n\n", err)
				}
			}
			return nil
		})
	if err != nil {
		fmt.Println(err)
	}
}

func parseFile(path string, standards *FairwindsStandards) error {
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
			if standards.IsOld(name, chart.Version) {
				color.Red("\t%s: %s\n", name, chart.Version)
			} else {
				fmt.Printf("\t%s: %s\n", name, chart.Version)
			}
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

// getStandards retrieves standards from s3
func getStandards() (*FairwindsStandards, error) {
	standards := FairwindsStandards{}
	response, err := http.Get(standardsURL)
	if err != nil {
		fmt.Printf("Error loading fairwinds standards: %v\n",err)
		return nil,err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Error reading fairwinds standards data from url: %v\n", err)
		return nil,err
	}

	err = yaml.Unmarshal(data, &standards)
	if err != nil {
		fmt.Printf("Error unmarshalling standards yaml: %v\n",err)
		return nil,err
	}
	return &standards,nil
}

func (s *FairwindsStandards) IsOld(release string, version string) bool {
	for chart, standardVersion := range s.Helm.Charts {
		if chart == release {
			return goversion.Compare(version,standardVersion.Version,"<")
		}
	}
	return false
}
