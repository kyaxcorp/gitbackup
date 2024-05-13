package main

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/spf13/afero"
)

// We have them here so that we can override these in the tests
var execCommand = exec.Command
var appFS = afero.NewOsFs()
var gitCommand = "git"
var archiveCommand = "/usr/bin/7z"
var gethomeDir = homedir.Dir

// Check if we have a copy of the repo already, if
// we do, we update the repo, else we do a fresh clone
func backUp(
	backupDir string,
	repo *Repository,
	bare bool,
	wg *sync.WaitGroup,
) ([]byte, error) {
	defer wg.Done()

	var dirName string
	if bare {
		dirName = repo.Name + ".git"
	} else {
		dirName = repo.Name
	}
	repoDir := path.Join(backupDir, repo.Namespace, dirName)

	_, err := appFS.Stat(repoDir)

	var stdoutStderr []byte
	if err == nil {
		log.Printf("%s exists, updating. \n", repo.Name)
		var cmd *exec.Cmd
		if bare {
			cmd = execCommand(gitCommand, "-C", repoDir, "remote", "update", "--prune")
		} else {
			cmd = execCommand(gitCommand, "-C", repoDir, "pull")
		}
		stdoutStderr, err = cmd.CombinedOutput()
	} else {
		log.Printf("Cloning %s\n", repo.Name)
		log.Printf("%#v\n", repo)

		if repo.Private && ignorePrivate != nil && *ignorePrivate {
			log.Printf("Skipping %s as it is a private repo.\n", repo.Name)
			return stdoutStderr, nil
		}

		if useHTTPSClone != nil && *useHTTPSClone {
			// Add username and token to the clone URL
			// https://gitlab.com/amitsaha/testproject1 => https://amitsaha:token@gitlab.com/amitsaha/testproject1
			u, err := url.Parse(repo.CloneURL)
			if err != nil {
				log.Fatalf("Invalid clone URL: %v\n", err)
			}
			repo.CloneURL = u.Scheme + "://" + gitHostUsername + ":" + gitHostToken + "@" + u.Host + u.Path
		}

		var cmd *exec.Cmd
		if bare {
			cmd = execCommand(gitCommand, "clone", "--mirror", repo.CloneURL, repoDir)
		} else {
			cmd = execCommand(gitCommand, "clone", repo.CloneURL, repoDir)
		}
		stdoutStderr, err = cmd.CombinedOutput()
	}

	if err != nil {
		return stdoutStderr, err
	}

	// Archive
	if appCfg.archiveDir != "" && err == nil {
		archiveArgs := []string{
			"a",
		}

		var suffix = ""
		if appCfg.archiveEncryptionPassword != "" {
			archiveArgs = append(archiveArgs, fmt.Sprintf("-p%s", appCfg.archiveEncryptionPassword))
			suffix = ".enc"
		}
		suffix += ".7z"

		archiveDirErr := os.MkdirAll(appCfg.archiveDir, 0751)
		if archiveDirErr != nil {
			return nil, archiveDirErr
		}

		now := time.Now()
		archiveDir := path.Join(appCfg.archiveDir, strings.Join([]string{repo.Namespace, dirName, now.Format("2006-01-02-15-04-05-0700")}, "-")+suffix)

		archiveArgs = append(archiveArgs, []string{
			"-v1500M",
			"t7z",
			"-m0=lzma2",
			"-mx=9",
			"-mfb=64",
			"-md=32m",
			"-ms=on",
			"-mhe=on",
			archiveDir,
			repoDir,
		}...)

		archiveCmd := execCommand(archiveCommand, archiveArgs...)
		archiveStdoutStderr, archiveErr := archiveCmd.CombinedOutput()

		if archiveErr != nil {
			return archiveStdoutStderr, archiveErr
		}
	}

	return stdoutStderr, err
}

func setupBackupDir(backupDir, service, githostURL *string) string {
	var gitHost, backupPath string
	var err error

	if len(*githostURL) != 0 {
		u, err := url.Parse(*githostURL)
		if err != nil {
			panic(err)
		}
		gitHost = u.Host
	} else {
		gitHost = knownServices[*service]
	}

	if len(*backupDir) == 0 {
		homeDir, err := gethomeDir()
		if err == nil {
			backupPath = path.Join(homeDir, ".gitbackup", gitHost)
		} else {
			log.Fatal("Could not determine home directory and backup directory not specified")
		}
	} else {
		backupPath = path.Join(*backupDir, gitHost)
	}

	err = createBackupRootDirIfRequired(backupPath)
	if err != nil {
		log.Fatalf("Error creating backup directory: %s %v", backupPath, err)
	}
	return backupPath
}

func createBackupRootDirIfRequired(backupPath string) error {
	return appFS.MkdirAll(backupPath, 0771)
}
