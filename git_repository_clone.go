package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const cacheSaveLastBackupDateAndContinueFromCache = "2006-01-02 15:04:05"

func getGithubSaveLastBackupDateAndContinueFromPath(c *appConfig) string {
	return filepath.Join(c.cacheDir, "github_save_last_backup_date_and_continue_from")
}

func handleGitRepositoryClone(client interface{}, c *appConfig) error {

	// Used for waiting for all the goroutines to finish before exiting
	startTime := time.Now()
	githubSaveLastBackupDateAndContinueFromCacheFilePath := getGithubSaveLastBackupDateAndContinueFromPath(c)
	debugLogf(
		"Starting backup run: service=%s backupDir=%s archiveDir=%s cacheDir=%s bare=%t maxConcurrentClones=%d useHTTPS=%t ignorePrivate=%t ignoreFork=%t shallowRepos=%d startFrom=%s saveLast=%t",
		c.service,
		c.backupDir,
		c.archiveDir,
		c.cacheDir,
		c.bare,
		c.maxConcurrentClones,
		c.useHTTPSClone,
		c.ignorePrivate,
		c.ignoreFork,
		len(c.shallowCloneRepos),
		c.githubStartFromLastPushAt,
		c.githubSaveLastBackupDateAndContinueFrom,
	)

	var wg sync.WaitGroup
	defer wg.Wait()

	tokens := make(chan bool, c.maxConcurrentClones)
	gitHostUsername = getUsername(client, c.service)

	if len(gitHostUsername) == 0 && !*ignorePrivate && *useHTTPSClone {
		log.Fatal("Your Git host's username is needed for backing up private repositories via HTTPS")
	}

	if c.githubSaveLastBackupDateAndContinueFrom {
		exists, err := fileExists(githubSaveLastBackupDateAndContinueFromCacheFilePath)
		if err != nil {
			return err
		}
		if exists {
			// load
			var tmpContent string
			tmpContent, err = getFileContents(githubSaveLastBackupDateAndContinueFromCacheFilePath)
			if tmpContent != "" {
				tmpTime, timeErr := time.Parse(cacheSaveLastBackupDateAndContinueFromCache, tmpContent)
				if timeErr != nil {
					// reset?! or report error...
					return timeErr
				}
				c.githubStartFromLastPushAt = tmpTime.Format(cacheSaveLastBackupDateAndContinueFromCache)
			}
		}
	}

	repositories, err := getRepositories(
		client,
		c,
		c.service,
		c.githubRepoType,
		c.githubNamespaceWhitelist,
		c.gitlabProjectVisibility,
		c.gitlabProjectMembershipType,
		c.ignoreFork,
	)

	if err != nil {
		return err
	}
	debugLogf("Retrieved %d repositories", len(repositories))

	for _, repo := range repositories {
		repo.Shallow = shallowCloneRequested(c.shallowCloneRepos, repo.Namespace, repo.Name)
		if repo.Shallow {
			debugLogf("Marked for shallow clone: %s/%s", repo.Namespace, repo.Name)
		}
	}

	if len(repositories) == 0 {
		return fmt.Errorf("no repositories retrieved")
	}

	isAnyErrorOccurred := false

	if c.githubSaveLastBackupDateAndContinueFrom {
		defer func() {
			// try saving the date...
			if isAnyErrorOccurred {
				log.Println("failed to cache the time of this current clone")
			} else {
				// Save the time
				cacheDirErr := os.MkdirAll(c.cacheDir, 0751)
				if cacheDirErr != nil {
					log.Println(fmt.Sprintf("failed to create cache dir -> %v", c.cacheDir))
					return
				}
				err = writeFileContents(githubSaveLastBackupDateAndContinueFromCacheFilePath, startTime.Format(cacheSaveLastBackupDateAndContinueFromCache))
				if err != nil {
					log.Println(fmt.Sprintf("failed to save cache file -> %s", githubSaveLastBackupDateAndContinueFromCacheFilePath))
				} else {
					log.Println(fmt.Sprintf("cache file saved successfully -> %s", githubSaveLastBackupDateAndContinueFromCacheFilePath))
				}
			}
		}()
	}

	log.Printf("Backing up %v repositories now..\n", len(repositories))
	for _, repo := range repositories {
		tokens <- true
		wg.Add(1)
		go func(repo *Repository) {
			debugLogf("Queueing repo: %s/%s (shallow=%t bare=%t)", repo.Namespace, repo.Name, repo.Shallow, c.bare)
			// Backup
			stdoutStderr, err := backUp(c.backupDir, repo, c.bare, &wg)
			if err != nil {
				isAnyErrorOccurred = true
				log.Printf("Error backing up %s: %s\n", repo.Name, stdoutStderr)
			}
			<-tokens
		}(repo)
	}
	return nil
}
