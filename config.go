package main

type appConfig struct {
	service                   string
	gitHostURL                string
	backupDir                 string
	archiveDir                string
	cacheDir                  string
	archiveEncryptionPassword string
	ignorePrivate             bool
	ignoreFork                bool
	debug                     bool
	useHTTPSClone             bool
	bare                      bool
	shallowCloneRepos         []string
	maxConcurrentClones       int

	// GitHub
	githubRepoType                    string
	githubNamespaceWhitelist          []string
	githubCreateUserMigration         bool
	githubCreateUserMigrationRetry    bool
	githubCreateUserMigrationRetryMax int
	githubListUserMigrations          bool
	githubWaitForMigrationComplete    bool
	//
	githubStartFromLastPushAt               string
	githubSaveLastBackupDateAndContinueFrom bool

	// Git Lab
	gitlabProjectVisibility     string
	gitlabProjectMembershipType string
}
