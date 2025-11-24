# gitbackup - Backup your GitHub, GitLab, and Bitbucket repositories

Code Quality [![Go Report Card](https://goreportcard.com/badge/github.com/kyaxcorp/gitbackup)](https://goreportcard.com/report/github.com/kyaxcorp/gitbackup)

## Introduction

``gitbackup`` is a tool to backup your git repositories from GitHub (including GitHub enterprise),
GitLab (including custom GitLab installations), or Bitbucket.

``gitbackup`` currently has two operation modes:

- The first and original operating mode is to create clones of only your git repository. This is supported for
  Bitbucket, GitHub and Gitlab.
- The second operating mode is only available for GitHub where you can create a user migration (including orgs) which
  you get back as a .tar.gz
  file containing all the artefacts that GitHub supports via their Migration API.

### Shallow clones (latest commit per branch)

If you only need the latest commit for specific repositories, pass `-shallow.repos` with a comma separated list of
`namespace/repo` names (works with bare and non-bare clones). Example: `-shallow.repos user1/repo1,org2/repo2`.

For those repos:
- Bare mode uses `git clone --mirror --depth=1 --no-single-branch`, then `git remote update --prune --depth=1 --no-tags`.
- Non-bare uses `git clone --depth=1 --no-single-branch`, then `git fetch origin --prune --depth=1 --no-tags`.

This keeps only the latest commit per branch. It is meant for backups only; the shallow mirror is not suitable for pushing.

## Running `gitbackup` from docker

```
docker pull gitbackup/gitbackup:latest
docker run \
--rm \
--name gitbackup \
-e GITHUB_TOKEN=$GITHUB_TOKEN \
-v /opt/gitbackup/backups:/gitbackup/backups \
-v /opt/gitbackup/archives:/gitbackup/archives \
-v /opt/gitbackup/cache:/gitbackup/cache \
gitbackup/gitbackup:latest \
-bare \
-maxConcurrentClones 1 \
-use-https-clone \
-service github \
-backupdir /gitbackup/backups \
-archive-dir /gitbackup/archives \
-cache-dir /gitbackup/cache \
-archive-encryption-password "1234567890" \
-github.startFromLastPushAt "2006-01-02 15:04:05" \
-github.saveLastBackupDateAndContinueFrom true
# optional: target only some repositories for shallow cloning
# -shallow.repos "user1/repo1,org2/repo2"

# Save your archives
...
```

## Using `gitbackup`

``gitbackup`` requires a [GitHub API access token](https://github.com/blog/1509-personal-api-tokens) for
backing up GitHub repositories, a [GitLab personal access token](https://gitlab.com/profile/personal_access_tokens)
for GitLab repositories, and a username and [app password](https://bitbucket.org/account/settings/app-passwords/) for
Bitbucket repositories.

You can supply the tokens to ``gitbackup`` using ``GITHUB_TOKEN`` and ``GITLAB_TOKEN`` environment variables
respectively, and the Bitbucket credentials with ``BITBUCKET_USERNAME`` and ``BITBUCKET_PASSWORD``.

### OAuth Scopes/Permissions required

#### Bitbucket

For the App password, the following permissions are required:

- `Account:Read`
- `Repositories:Read`

#### GitHub

- `repo`: Reading repositories, including private repositories
- `user` and `admin:org`: Basically, this gives `gitbackup` a lot of permissions than you may be comfortable with.
  However, these are required for the user migration and org migration operations.

#### GitLab

- `api`: Grants complete read/write access to the API, including all groups and projects.
  For some reason, `read_user` and `read_repository` is not sufficient.

### Security and credentials

When you provide the tokens via environment variables, they remain accessible in your shell history
and via the processes' environment for the lifetime of the process. By default, SSH authentication
is used to clone your repositories. If `use-https-clone` is specified, private repositories
are cloned via `https` basic auth and the token provided will be stored in the repositories'
`.git/config`.

### GitBackup Help

Typing ``-help`` will display the command line options that `gitbackup` recognizes:

```
$ gitbackup -help
Usage of ./gitbackup:
  -archive-dir string
        Backup Archive directory
  -archive-encryption-password string
        Archive Encryption Password
  -backupdir string
        Backup directory
  -bare
        Clone bare repositories
  -cache-dir string
        Cache directory
  -githost.url string
        DNS of the custom Git host
  -github.createUserMigration
        Download user data
  -github.createUserMigrationRetry
        Retry creating the GitHub user migration if we get an error (default true)
  -github.createUserMigrationRetryMax int
        Number of retries to attempt for creating GitHub user migration (default 5)
  -github.listUserMigrations
        List available user migrations
  -github.namespaceWhitelist string
        Organizations/Users from where we should clone (separate each value by a comma: 'user1,org2')
  -github.repoType string
        Repo types to backup (all, owner, member, starred) (default "all")
  -github.saveLastBackupDateAndContinueFrom
        Backup only from the last clone datetime when a full successful backup of all repositories was complete, it can be used with github.startFromLastPushAt and it will be ignored after (default true)
  -github.startFromLastPushAt string
        Start backing up the repo which has a Push Equal or Higher than specified
  -github.waitForUserMigration
        Wait for migration to complete (default true)
  -gitlab.projectMembershipType string
        Project type to clone (all, owner, member, starred) (default "all")
  -gitlab.projectVisibility string
        Visibility level of Projects to clone (internal, public, private) (default "internal")
  -ignore-fork
        Ignore repositories which are forks
  -ignore-private
        Ignore private repositories/projects
  -maxConcurrentClones int
        Max Number of Concurrent Clones (default 10)
  -service string
        Git Hosted Service Name (github/gitlab/bitbucket)
  -shallow.repos string
        Comma separated full repo names (namespace/name) to shallow clone (latest commit per branch)
  -use-https-clone
        Use HTTPS for cloning instead of SSH
```
