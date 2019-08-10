package cmd

/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add directory into an existing structure year/month",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		addFunction()
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().BoolVarP(&sortByTime, "sort", "s", false, "Sort files by creation date")
	addCmd.Flags().StringVarP(&inputDirectory, "input", "i", "", "add this directory to current structure (required)")
	addCmd.Flags().StringVarP(&outputDirectory, "output", "o", "Final", "this directory is the current structure year/month (required)")

	addCmd.MarkFlagRequired("input")
}

func addFunction() {
	var (
		filesList    *[]string
		allFilesList *[]string
	)

	upperCaseAllExtensions()

	if recursiveSearch {
		filesList, allFilesList = getFilesListWithRecursiveSearch()
	} else {
		filesList, allFilesList = getFilesListWithoutRecursiveSearch("")
	}

	showExtension(filesList, allFilesList)
	showAddAction()
	if !skipWarning {
		showDanger()
	}
	moveFiles(filesList)
	pathList := getPathToRemoveDuplicates(filesList)

	sort.Strings(*pathList)

	for num, p := range *pathList {
		fmt.Printf("In \"%s\" (%d/%d): \n", p, num+1, len(*pathList))

		_, filesList := getFilesListWithoutRecursiveSearch(p)
		remainingFilesList := removeDuplicates(filesList)

		if sortByTime {
			removeNumberInFrontFilename(remainingFilesList)
			sortListByTime(remainingFilesList)
		}
	}
}

func getPathToRemoveDuplicates(filesList *[]string) *[]string {
	pathMap := make(map[string]bool)

	for _, filename := range *filesList {
		pathMap[filepath.Dir(filename)] = true
	}

	pathList := make([]string, len(pathMap))
	i := 0

	for p := range pathMap {
		pathList[i] = p
		i++
	}

	return &pathList
}

func showAddAction() {
	if recursiveSearch {
		fmt.Println("Recursive search (-r):                  [\x1b[32;1mActivated\x1b[0m]")
	} else {
		fmt.Println("Recursive search (-r):                  [\x1b[31;1mDisabled\x1b[0m]")
	}

	if skipWarning {
		fmt.Println("skip warning (-k):                      [\x1b[32;1mActivated\x1b[0m]")
	} else {
		fmt.Println("skip warning (-k):                      [\x1b[31;1mDisabled\x1b[0m]")
	}

	if sortByTime {
		fmt.Println("Sort files by time (-s):                [\x1b[32;1mActivated\x1b[0m]")
	} else {
		fmt.Println("Sort files by time (-s):                [\x1b[31;1mDisabled\x1b[0m]")
	}

	fmt.Printf("Input directory (-i):                   \"%s\"\n", inputDirectory)

	fmt.Printf("Output directory (-o):                  \"%s\"\n", outputDirectory)
}
