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
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	recursiveSearch  bool
	skipWarning      bool
	sortByTime       bool
	cfgFile          string
	inputDirectory   string
	outputDirectory  string
	wantedExtensions string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ranking",
	Short: "It move pictures/movies (by default) into year/month directories",
	Long: `It move, search duplicate, sort pictures/movies into a structure
year/month directories`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ranking.yaml)")

	rootCmd.PersistentFlags().BoolVarP(&recursiveSearch, "recursive", "r", false, "Recursive search")
	rootCmd.PersistentFlags().BoolVarP(&skipWarning, "skip", "k", false, "Skip warning message")
	rootCmd.PersistentFlags().StringVarP(&wantedExtensions, "wantedExtensions", "w", "*.jpg *.jpeg *.png *.svg *.gif *.avi *.mp4 *.mkv *.mts *.nef *.odt *.vob *.mpg *.wmv *.3gp *.ppt *.mov *.pdf *.bmp *.zip *.xcf *.tiff", "Wanted extensions")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".ranking" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".ranking")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func changeName(filename *string) {
	if !isFileExists(filename) {
		return
	}

	filenameWithoutExt := strings.TrimSuffix(*filename, filepath.Ext(*filename))
	ext := filepath.Ext(*filename)
	i := 0

	for {
		tmpFilename := filenameWithoutExt + "-" + strconv.Itoa(i) + ext

		if !isFileExists(&tmpFilename) {
			*filename = tmpFilename

			return
		}

		i++
	}
}

// Search files in the current directory only. Not in subfolder
// Return a pointer to a list of string
func getFilesListWithoutRecursiveSearch(myPath string) (*[]string, *[]string) {
	filesList := []string{}
	extensionsSliptted := strings.Split(wantedExtensions, " ")
	searchGlob := "*"

	if myPath != "" {
		searchGlob = filepath.Join(myPath, searchGlob)
	} else if inputDirectory != "" {
		searchGlob = filepath.Join(inputDirectory, searchGlob)
	}

	allFilesList, _ := filepath.Glob(searchGlob)

	for _, ext := range extensionsSliptted {
		files := []string{}
		if myPath != "" {
			files, _ = filepath.Glob(filepath.Join(myPath, ext))
		} else if inputDirectory != "" {
			files, _ = filepath.Glob(filepath.Join(inputDirectory, ext))
		} else {
			files, _ = filepath.Glob(ext)
		}

		for _, f := range files {
			info, err := os.Stat(f)
			if err != nil {
				log.Printf("\nIn \"getFilesListWithoutRecursiveSearch\" func, os.Stat had an issue: %s\n", err)
			}
			if !info.IsDir() {
				filesList = append(filesList, f)
			}
		}
	}

	return &filesList, &allFilesList
}

// Search files in the current directory and subfolders
// Return a pointer to a list of string
func getFilesListWithRecursiveSearch() (*[]string, *[]string) {
	filesList := []string{}
	allFilesList := []string{}
	extensionsSliptted := strings.Split(wantedExtensions, " ")
	searchWalk := "."

	if inputDirectory != "" {
		searchWalk = filepath.Join(searchWalk, inputDirectory)
	}

	err := filepath.Walk(searchWalk, func(fp string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}

		extFile := filepath.Ext(fp)
		if extFile == "" || extFile == "." {
			return nil
		}

		allFilesList = append(allFilesList, fp)

		for _, ext := range extensionsSliptted {
			if strings.Contains(ext, extFile) {
				filesList = append(filesList, fp)
				break
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("\nIn \"getFilesListWithRecursiveSearch\" func, filepath.walk had an issue: %s\n", err)
	}

	return &filesList, &allFilesList
}

func getExifDateFilename(filename *string) (time.Time, error) {
	f, err := os.Open(*filename)
	if err != nil {
		return time.Now(), err
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return getModTime(filename)
	}

	tm, err := x.DateTime()
	if err != nil {
		return getModTime(filename)
	}

	return tm, nil
}

// Calculate the md5sum of file
// Return a pointer to a list of byte
func getMd5(filenamePtr *string) []byte {
	file, err := os.Open(*filenamePtr)
	if err != nil {
		log.Printf("\nIn \"getMd5\" func, os.Open had an issue: %s\n", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		log.Printf("\nIn \"getMd5\" func, io.Copy had an issue: %s\n", err)
	}

	return hash.Sum(nil)
}

func getModTime(filename *string) (time.Time, error) {
	info, err := os.Stat(*filename)
	if err != nil {
		return time.Now(), err
	}

	return info.ModTime(), nil
}

func isFileExists(filename *string) bool {
	if _, err := os.Stat(*filename); os.IsNotExist(err) {
		return false
	}

	return true
}

// Move files into respective folder depending when file where created
func moveFiles(filesListPtr *[]string) {
	root, _ := os.Getwd()
	totalFiles := len(*filesListPtr)
	monthMap := map[int8]string{
		1:  "1 - Janvier",
		2:  "2 - Février",
		3:  "3 - Mars",
		4:  "4 - Avril",
		5:  "5 - Mai",
		6:  "6 - Juin",
		7:  "7 - Juillet",
		8:  "8 - Août",
		9:  "9 - Septembre",
		10: "10 - Octobre",
		11: "11 - Novembre",
		12: "12 - Décembre",
	}

	for num, filename := range *filesListPtr {
		fmt.Printf("Moving files: %d/%d\r", num+1, totalFiles)

		date, err := getExifDateFilename(&filename)
		if err != nil {
			log.Printf("\nIn \"moveFiles\" func, os.Stat had an issue: %s\n", err)
			continue
		}

		myFilePath := path.Join(root, outputDirectory, strconv.Itoa(date.Year()), monthMap[int8(date.Month())])
		if _, err = os.Stat(myFilePath); err != nil {
			if err = os.MkdirAll(myFilePath, 0744); err != nil {
				log.Printf("\nIn \"moveFiles\" func, os.MkdirAll had an issue: %s\n", err)
			}
		}
		_, nameFile := path.Split(filename)
		newLocation := path.Join(myFilePath, nameFile)

		changeName(&newLocation)

		if err = os.Rename(filename, newLocation); err != nil {
			log.Printf("\nIn \"moveFiles\" func, os.Rename had an issue: %s\n", err)
		}

		(*filesListPtr)[num] = newLocation
	}

	fmt.Println()
}

// Remove duplicate files
// Return a pointer to a list of string. Contains list of remaining files
func removeDuplicates(filesList *[]string) *[]string {
	totalFiles := len(*filesList)
	md5Map := make(map[string]string, totalFiles)
	duplicatesList := []string{}

	for num, filename := range *filesList {
		fmt.Printf("Search duplicates: %v/%v | Found: %v\r", num+1, totalFiles, len(duplicatesList))
		md5sumStr := fmt.Sprintf("%x", getMd5(&filename))

		if _, ok := md5Map[md5sumStr]; !ok {
			md5Map[md5sumStr] = filename
		} else {
			duplicatesList = append(duplicatesList, filename)
		}
	}

	fmt.Println()

	totalFiles = len(duplicatesList)
	for num, filename := range duplicatesList {
		fmt.Printf("Remove duplicate %v/%v\r", num+1, totalFiles)
		os.Remove(filename)
	}

	if len(duplicatesList) > 0 {
		fmt.Println()
	}

	remainingFiles := make([]string, len(md5Map))
	i := 0
	for _, val := range md5Map {
		remainingFiles[i] = val
		i++
	}

	return &remainingFiles
}

func removeNumberInFrontFilename(filesList *[]string) {
	numberRegex := regexp.MustCompile("^\\d+ - ")

	for num, fn := range *filesList {
		fmt.Printf("Remove number in filename's front: %d/%d\r", num+1, len(*filesList))
		filename := filepath.Base(fn)

		if numberRegex.MatchString(filename) {
			root := filepath.Dir(fn)

			for numberRegex.MatchString(filename) {
				filename = numberRegex.ReplaceAllString(filename, "")
			}

			if len(filename) == 0 {
				log.Printf("\nWo, \"%s\" become an empty filename. Continue without change !\n", fn)
			}

			newFilename := filepath.Join(root, filename)

			if isFileExists(&newFilename) {
				ext := filepath.Ext(newFilename)
				i := 0

				for {
					filenameWithoutExt := newFilename[0 : len(newFilename)-len(ext)]
					tmpName := filenameWithoutExt + "-" + strconv.Itoa(i) + ext

					if !isFileExists(&tmpName) {
						newFilename = tmpName
						break
					}
					i++
				}
			}

			if err := os.Rename(fn, newFilename); err != nil {
				log.Printf("\nIn \"moveFiles\" func, os.Rename had an issue: %s\n", err)
			} else {
				(*filesList)[num] = newFilename
			}
		}
	}
	fmt.Println()
}

func showExtension(filesList *[]string, allFilesList *[]string) {
	fmt.Printf("Search extensions: %s\n\n", wantedExtensions)

	// Extensions found
	extensionsMap := make(map[string]bool)
	for _, filename := range *filesList {
		extensionsMap[filepath.Ext(filename)] = true
	}
	extensionsList := make([]string, len(extensionsMap))
	i := 0
	for k := range extensionsMap {
		extensionsList[i] = k
		i++
	}

	if len(extensionsList) == 0 {
		fmt.Println("Founded extensions: no extension founded\nNothing to do !")
		os.Exit(0)
	}

	sort.Strings(extensionsList)

	fmt.Printf("Founded extensions: %s\n\n", strings.Join(extensionsList, " "))

	// Unmanaged extensions
	allExtensionsMap := make(map[string]bool)
	for _, filename := range *allFilesList {
		allExtensionsMap[filepath.Ext(filename)] = false
	}

	for _, ext := range extensionsList {
		allExtensionsMap[ext] = true
	}

	deltaList := []string{}
	for k, v := range allExtensionsMap {
		if !v {
			deltaList = append(deltaList, k)
		}
	}

	if len(deltaList) != 0 {
		if len(deltaList) > 1 {
			sort.Strings(deltaList)
		}
		fmt.Printf("Extensions not taken into account: %s\n\n", strings.Join(deltaList, " "))
	} else {
		fmt.Printf("Extensions not taken into account: not found\n\n")
	}
}

// Show this program might be dangerous. Can break filesystem
func showDanger() {
	if skipWarning {
		return
	}

	fmt.Println("\x1b[36;1mBe carefull. You can break your system !\x1b[0m")
	fmt.Print("Starting in (ctrl+c to stop):")
	for i := 10; i > 0; i-- {
		fmt.Print(fmt.Sprintf(" \x1b[31;1m%d\x1b[0m", i))
		time.Sleep(1 * time.Second)
	}
	fmt.Println()
}

// Adding number before filename after having sorted files
func sortListByTime(filesList *[]string) {
	tmpFilesList := []thing{}
	filesNumber := len(*filesList)

	for num, filename := range *filesList {
		date, err := getExifDateFilename(&filename)
		if err != nil {
			log.Printf("\nIn \"sortByTime\" func, os.Stat had an issue: %s\n", err)
			continue
		}
		t := thing{
			date:     date,
			filename: filepath.Base(filename),
			path:     filepath.Dir(filename),
			remove:   false,
		}
		tmpFilesList = append(tmpFilesList, t)

		fmt.Printf("Getting files date creation: %d/%d\r", num+1, filesNumber)
	}

	fmt.Println()

	count := 1
	filesNumber = len(tmpFilesList)

	for {
		if len(tmpFilesList) == 0 {
			break
		}

		myFilePath := tmpFilesList[0].path
		tmpFilesList[0].remove = true
		sameLocation := []thing{}

		for num, t := range tmpFilesList {
			if t.path == myFilePath {
				tmpFilesList[num].remove = true
				sameLocation = append(sameLocation, t)
			}
		}

		// Remove removable for slice
		for i := 0; i < len(tmpFilesList); i++ {
			if tmpFilesList[i].remove {
				tmpFilesList = append(tmpFilesList[:i], tmpFilesList[i+1:]...)
				i--
			}
		}

		// Sort slice by time
		sort.Sort(byDate(sameLocation))

		// Rename item in slice
		for num, t := range sameLocation {
			oldLocation := filepath.Join(t.path, t.filename)
			newLocation := filepath.Join(t.path, strconv.Itoa(num+1)+" - "+t.filename)

			fmt.Printf("Rename files: %d/%d\r", count, filesNumber)

			count++
			os.Rename(oldLocation, newLocation)
		}
	}
	fmt.Println()
}

func upperCaseAllExtensions() {
	wantedExtensionsSplitted := strings.Split(wantedExtensions, " ")

	wantedExtensionsMap := make(map[string]bool, len(wantedExtensions)*2)

	for _, ext := range wantedExtensionsSplitted {
		wantedExtensionsMap[ext] = true
		wantedExtensionsMap[strings.ToUpper(ext)] = true
	}

	i := 0
	wantedExtensionsKeys := make([]string, len(wantedExtensionsMap))
	for k := range wantedExtensionsMap {
		wantedExtensionsKeys[i] = k
		i++
	}

	sort.Strings(wantedExtensionsKeys)

	wantedExtensions = strings.Join(wantedExtensionsKeys, " ")
}
