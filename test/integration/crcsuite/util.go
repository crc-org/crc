package crcsuite

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	clicumber "github.com/code-ready/clicumber/testsuite"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Delete CRC instance
func DeleteCRC() error {

	command := "crc delete"
	_ = clicumber.ExecuteCommand(command)

	fmt.Printf("Deleted CRC instance (if one existed).\n")
	return nil
}

// Remove CRC home folder ~/.crc
func RemoveCRCHome() error {

	keepFile := filepath.Join(CRCHome, ".keep")

	_, err := os.Stat(keepFile)
	if err != nil { // cannot get keepFile's status
		err = os.RemoveAll(CRCHome)

		if err != nil {
			fmt.Printf("Problem deleting CRC home folder %s.\n", CRCHome)
			return err
		}

		fmt.Printf("Deleted CRC home folder %s.\n", CRCHome)
		return nil

	}
	// keepFile exists
	return fmt.Errorf("Folder %s not removed as per request: %s present", CRCHome, keepFile)
}

// Post a file to Github repo via Github API
// localFile:     absolute path to file to post
// githubRepo:    repository on Github that will host the file
// githubFile:    filename of the file on Github
// commitMessage: commit message to accompany the file
func CreateGithubFile(localFile string, githubRepository string, githubFile string, commitMessage string) error {

	usr, _ := user.Current()
	ctx := context.Background()

	// authenticate & create client
	tokenBytes, err := ioutil.ReadFile(filepath.Join(usr.HomeDir, ".ssh", "crc-data-gh-token"))
	if err != nil {
		fmt.Printf("Cannot read the token: %s", err)
		return err
	}
	tokenString := string(tokenBytes)
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: tokenString},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// prepare fileContent to be posted to Github
	fileContent, err := ioutil.ReadFile(localFile)
	if err != nil {
		fmt.Println("Could not read local file:", err)
		return err
	}

	// the file must not already exist on Github
	// specify options
	opts := &github.RepositoryContentFileOptions{
		Message:   github.String(commitMessage),
		Content:   fileContent,
		Branch:    github.String("master"),
		Committer: &github.CommitAuthor{Name: github.String("Jakub Sliacan"), Email: github.String("jsliacan@redhat.com")},
	}

	_, _, err = client.Repositories.CreateFile(ctx, "jsliacan", githubRepository, githubFile, opts)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
