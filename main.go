package main

import (
	"os/exec"
	"bytes"
	"log"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"regexp"
)

type Branch struct {
	Name string
	IsMerged bool
	IsOutdated bool
	Author string
	LastUpdated string
}

func runGitCommand(args []string) string {
	cmd := exec.Command("git", args...)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Command failed %s : %s", strings.Join(args, " "), err)
	}
	return out.String()
}

func clone(repoName string) string {
	tempDir, errTmp := ioutil.TempDir(os.TempDir(), "branchitto")

	if errTmp != nil {
		log.Fatal(errTmp)
	}

	runGitCommand([]string{"clone", "-q", repoName, tempDir})
	return tempDir
}

func getMerged(tmpDir string) []Branch {
	master := regexp.MustCompile("(origin/HEAD|origin/master)")
	errChdir := os.Chdir(tmpDir)
	if errChdir != nil {
		log.Fatalf("Can't change folder to %s", tmpDir)
	}

	branches := make([]Branch, 0)
	merged := runGitCommand([]string{"branch", "-r", "--merged"})
	for _, branchName := range strings.Split(merged, "\n") {
		if master.MatchString(strings.TrimSpace(branchName)) { continue }
		if len(branchName) == 0 { continue }
		branches = append(branches, getBranchData(strings.TrimSpace(branchName), true, false))
	}

	notMerged := runGitCommand([]string{"branch", "-r", "--no-merged"})
	for _, branchName := range strings.Split(notMerged, "\n") {
		if len(branchName) == 0 { continue }
		logLastMonth := runGitCommand([]string{"log", "-1", "--since='1 month ago'", "-s", strings.TrimSpace(branchName), "--oneline"})
		isOutdated := len(logLastMonth) == 0 
		branches = append(branches, getBranchData(strings.TrimSpace(branchName), false, isOutdated))
	}

	return branches
}

func getBranchData (branchName string, isMerged, isOutdated bool) Branch {
	info := runGitCommand([]string{"show", "--format=\"%ci,%an,%cn\"", branchName})
	info2 := strings.Split(strings.Split(info, "\n")[0], ",")
	return Branch{
		branchName,
		isMerged,
		isOutdated,
		info2[1],
		info2[0],
	}
}

func main() {

	folder := clone("https://github.com/mzabriskie/axios")
	fmt.Println(folder)

	merged := getMerged(folder)

	for _, b := range merged {
		log.Printf("branch: %s - %s - %s", b.Name, b.LastUpdated, b.Author)
	}
}