package main

import (
	"errors"
	"flag"
	"strings"
)

var appCfg appConfig

func initConfig(args []string) (*appConfig, error) {

	var githubNamespaceWhitelistString string
	var shallowCloneReposString string

	fs := flag.NewFlagSet("gitbackup", flag.ExitOnError)

	// Generic flags
	fs.StringVar(&appCfg.service, "service", "", "Git Hosted Service Name (github/gitlab/bitbucket)")
	fs.StringVar(&appCfg.gitHostURL, "githost.url", "", "DNS of the custom Git host")
	fs.StringVar(&appCfg.backupDir, "backupdir", "", "Backup directory")
	fs.StringVar(&appCfg.archiveDir, "archive-dir", "", "Backup Archive directory")
	fs.StringVar(&appCfg.cacheDir, "cache-dir", "", "Cache directory")
	fs.StringVar(&appCfg.archiveEncryptionPassword, "archive-encryption-password", "", "Archive Encryption Password")
	fs.BoolVar(&appCfg.ignorePrivate, "ignore-private", false, "Ignore private repositories/projects")
	fs.BoolVar(&appCfg.ignoreFork, "ignore-fork", false, "Ignore repositories which are forks")
	fs.BoolVar(&appCfg.debug, "debug", false, "Enable verbose debug logging")
	fs.BoolVar(&appCfg.useHTTPSClone, "use-https-clone", false, "Use HTTPS for cloning instead of SSH")
	fs.BoolVar(&appCfg.bare, "bare", false, "Clone bare repositories")
	fs.StringVar(&shallowCloneReposString, "shallow.repos", "", "Comma separated full repo names (namespace/name) to shallow clone (latest commit per branch)")

	// GitHub specific flags
	fs.StringVar(&appCfg.githubRepoType, "github.repoType", "all", "Repo types to backup (all, owner, member, starred)")

	fs.StringVar(&appCfg.githubStartFromLastPushAt,
		"github.startFromLastPushAt",
		"",
		"Start backing up the repo which has a Push Equal or Higher than specified",
	)

	fs.BoolVar(&appCfg.githubSaveLastBackupDateAndContinueFrom,
		"github.saveLastBackupDateAndContinueFrom",
		true,
		"Backup only from the last clone datetime when a full successful backup of all repositories was complete, it can be used with github.startFromLastPushAt and it will be ignored after",
	)

	fs.StringVar(
		&githubNamespaceWhitelistString, "github.namespaceWhitelist",
		"", "Organizations/Users from where we should clone (separate each value by a comma: 'user1,org2')",
	)
	fs.BoolVar(&appCfg.githubCreateUserMigration, "github.createUserMigration", false, "Download user data")
	fs.BoolVar(
		&appCfg.githubCreateUserMigrationRetry, "github.createUserMigrationRetry", true,
		"Retry creating the GitHub user migration if we get an error",
	)
	fs.IntVar(
		&appCfg.githubCreateUserMigrationRetryMax, "github.createUserMigrationRetryMax",
		defaultMaxUserMigrationRetry,
		"Number of retries to attempt for creating GitHub user migration",
	)
	fs.IntVar(
		&appCfg.maxConcurrentClones, "maxConcurrentClones",
		10,
		"Max Number of Concurrent Clones",
	)
	fs.BoolVar(
		&appCfg.githubListUserMigrations,
		"github.listUserMigrations",
		false,
		"List available user migrations",
	)
	fs.BoolVar(
		&appCfg.githubWaitForMigrationComplete,
		"github.waitForUserMigration",
		true,
		"Wait for migration to complete",
	)

	// Gitlab specific flags
	fs.StringVar(
		&appCfg.gitlabProjectVisibility,
		"gitlab.projectVisibility",
		"internal",
		"Visibility level of Projects to clone (internal, public, private)",
	)
	fs.StringVar(
		&appCfg.gitlabProjectMembershipType,
		"gitlab.projectMembershipType", "all",
		"Project type to clone (all, owner, member, starred)",
	)

	err := fs.Parse(args)
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		return nil, err
	}

	useHTTPSClone = &appCfg.useHTTPSClone
	ignorePrivate = &appCfg.ignorePrivate

	// Split namespaces
	if len(appCfg.githubNamespaceWhitelist) > 0 {
		appCfg.githubNamespaceWhitelist = strings.Split(githubNamespaceWhitelistString, ",")
	}
	if len(shallowCloneReposString) > 0 {
		appCfg.shallowCloneRepos = strings.Split(shallowCloneReposString, ",")
	}
	appCfg.backupDir = setupBackupDir(&appCfg.backupDir, &appCfg.service, &appCfg.gitHostURL)
	return &appCfg, nil
}

func validateConfig(c *appConfig) error {
	if _, ok := knownServices[c.service]; !ok {
		return errors.New("Please specify the git service type: github, gitlab, bitbucket")
	}

	if !validGitlabProjectMembership(c.gitlabProjectMembershipType) {
		return errors.New("Please specify a valid gitlab project membership - all/owner/member")
	}
	return nil
}
