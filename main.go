package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"strings"
)

func main() {
	var folder string
	var email string

	flag.StringVar(&folder, "add", "", "add a folder to the scanning list")
	flag.StringVar(&email, "email", "email@email.com", "add the email to the scan")
	flag.Parse()

	if folder != "" {
		scan(folder)
	}

	stats(email)
}

func scan(folder string) {
	fmt.Printf("found folders\n\n")
	repos := recursiveScanFolder(folder)
	filepath := getDotFilepath()
	addNewSliceElementsToFolder(repos, filepath)
	fmt.Printf("\n\nsuccesfully added\n\n")
}

func stats(email string) {
	fmt.Printf("stats %s\n", email)
}

func getDotFilepath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	return usr.HomeDir + "/.gogitlocalstats"
}

func recursiveScanFolder(folder string) []string {
	return scanGitFolders([]string{}, folder)
}

func addNewSliceElementsToFolder(repos []string, filepath string) {
	existingRepos := parseFileLinesToSlice(filepath)
	newRepos := joinSlices(existingRepos, repos)
	dumpStringsSliceToFile(newRepos, filepath)
}

func dumpStringsSliceToFile(repos []string, filepath string) {
	content := strings.Join(repos, "\n")
	err := os.WriteFile(filepath, []byte(content), 0755)
	if err != nil {
		log.Fatal(err)
	}
}

func joinSlices(newList []string, oldList []string) []string {
	oldMap := map[string]bool{}
	for _, key := range oldList {
		oldMap[key] = true
	}

	for _, path := range newList {
		if _, ok := oldMap[path]; !ok {
			oldList = append(oldList, path)
		}
	}

	return oldList
}

func parseFileLinesToSlice(filepath string) []string {
	file := openFile(filepath)
	defer file.Close()

	var lines []string

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			fmt.Println("this?")
			panic(err)
		}
	}

	return lines
}

func openFile(filePath string) *os.File {
	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			f, err = os.Create(filePath)
			if err != nil {
				panic(fmt.Errorf("failed to create file: %w", err))
			}
			f.Close()
			f, err = os.Open(filePath)
			if err != nil {
				panic(fmt.Errorf("failed to reopen file: %w", err))
			}
		} else {
			panic(fmt.Errorf("failed to open file: %w", err))
		}
	}
	return f
}

func scanGitFolders(folders []string, folder string) []string {
	folder = strings.TrimSuffix(folder, "/")

	file, err := os.Open(folder)
	if err != nil {
		log.Fatal(err)
	}

	dirs, err := file.ReadDir(-1)
	file.Close()
	if err != nil {
		log.Fatal(err)
	}

	var path string

	for _, dir := range dirs {
		if dir.IsDir() {
			path = folder + "/" + dir.Name()

			if dir.Name() == ".git" {
				path = strings.TrimSuffix(path, "/.git")
				folders = append(folders, path)
				fmt.Println(path)
				continue
			}
			if dir.Name() == "vendor" || dir.Name() == "node_modules" {
				continue
			}
			folders = scanGitFolders(folders, path)
		}
	}

	return folders
}
