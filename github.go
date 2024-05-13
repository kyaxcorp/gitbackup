package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/v34/github"
)

func getGithubRepositories(
	client interface{},
	c *appConfig,
	service string,
	githubRepoType string,
	githubNamespaceWhitelist []string,
	gitlabProjectVisibility string,
	gitlabProjectMembershipType string,
	ignoreFork bool,
) ([]*Repository, error) {

	if client == nil {
		log.Fatalf("Couldn't acquire a client to talk to %s", service)
	}

	var repositories []*Repository
	var cloneURL string

	ctx := context.Background()

	var err error
	var startFromLastPushAt time.Time
	var startFromLastPush bool

	if c.githubStartFromLastPushAt != "" {
		startFromLastPush = true
		startFromLastPushAt, err = time.Parse(cacheSaveLastBackupDateAndContinueFromCache, c.githubStartFromLastPushAt)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed to parse githubStartFromLastPushAt -> %v", err.Error()))
		}
	}

	if githubRepoType == "starred" {
		options := github.ActivityListStarredOptions{}
		for {
			stars, resp, err := client.(*github.Client).Activity.ListStarred(ctx, "", &options)
			if err == nil {
				for _, star := range stars {
					if *star.Repository.Fork && ignoreFork {
						continue
					}
					namespace := strings.Split(*star.Repository.FullName, "/")[0]
					if useHTTPSClone != nil && *useHTTPSClone {
						cloneURL = *star.Repository.CloneURL
					} else {
						cloneURL = *star.Repository.SSHURL
					}

					if startFromLastPush {
						if star.Repository.PushedAt != nil {
							if !star.Repository.PushedAt.Time.After(startFromLastPushAt) {
								continue
							}
						}
					}

					repositories = append(repositories, &Repository{
						PushedAt:  star.Repository.PushedAt,
						UpdatedAt: star.Repository.UpdatedAt,
						//
						CloneURL:  cloneURL,
						Name:      *star.Repository.Name,
						Namespace: namespace,
						Private:   *star.Repository.Private,
					})
				}
			} else {
				return nil, err
			}
			if resp.NextPage == 0 {
				break
			}
			options.ListOptions.Page = resp.NextPage
		}
		return repositories, nil
	}

	options := github.RepositoryListOptions{Type: githubRepoType}
	githubNamespaceWhitelistLength := len(githubNamespaceWhitelist)

	for {
		repos, resp, err := client.(*github.Client).Repositories.List(ctx, "", &options)
		if err == nil {
			for _, repo := range repos {
				if *repo.Fork && ignoreFork {
					continue
				}

				namespace := strings.Split(*repo.FullName, "/")[0]

				if githubNamespaceWhitelistLength > 0 && !contains(githubNamespaceWhitelist, namespace) {
					continue
				}

				if useHTTPSClone != nil && *useHTTPSClone {
					cloneURL = *repo.CloneURL
				} else {
					cloneURL = *repo.SSHURL
				}

				if startFromLastPush {
					if repo.PushedAt != nil {
						if !repo.PushedAt.Time.After(startFromLastPushAt) {
							continue
						}
					}
				}

				repositories = append(repositories, &Repository{
					PushedAt:  repo.PushedAt,
					UpdatedAt: repo.UpdatedAt,
					CloneURL:  cloneURL,
					Name:      *repo.Name,
					Namespace: namespace,
					Private:   *repo.Private,
				})
			}
		} else {
			return nil, err
		}
		if resp.NextPage == 0 {
			break
		}
		options.ListOptions.Page = resp.NextPage
	}
	return repositories, nil
}
