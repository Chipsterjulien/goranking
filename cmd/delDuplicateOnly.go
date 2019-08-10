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

	"github.com/spf13/cobra"
)

// delOnlyCmd represents the delOnly command
var delOnlyCmd = &cobra.Command{
	Use:   "delDupOnly",
	Short: "Only search duplicates files to delete them",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		delDupOnlyFunction()
	},
}

func init() {
	rootCmd.AddCommand(delOnlyCmd)

	delOnlyCmd.Flags().StringVarP(&inputDirectory, "input", "i", "", "Path to files which will be moved")
}

func delDupOnlyFunction() {
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
	showDelDuplicateOnlyAction()

	if !skipWarning {
		showDanger()
	}

	removeDuplicates(filesList)
}

func showDelDuplicateOnlyAction() {
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

	fmt.Printf("Input directory (-i):                   \"%s\"\n", inputDirectory)

}
