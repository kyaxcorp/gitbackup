package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v34/github"
	"github.com/ktrysmt/go-bitbucket"
	gitlab "github.com/xanzy/go-gitlab"
)

func getUsername(client interface{}, service string) string {

	if client == nil {
		log.Fatalf("Couldn't acquire a client to talk to %s", service)
	}

	if service == "github" {
		ctx := context.Background()
		user, _, err := client.(*github.Client).Users.Get(ctx, "")
		if err != nil {
			log.Fatal("Error retrieving username", err.Error())
		}
		return *user.Login
	}

	if service == "gitlab" {
		user, _, err := client.(*gitlab.Client).Users.CurrentUser()
		if err != nil {
			log.Fatal("Error retrieving username", err.Error())
		}
		return user.Username
	}

	if service == "bitbucket" {
		user, err := client.(*bitbucket.Client).User.Profile()
		if err != nil {
			log.Fatal("Error retrieving username", err.Error())
		}
		return user.Username
	}

	return ""
}

func validGitlabProjectMembership(membership string) bool {
	validMemberships := []string{"all", "owner", "member", "starred"}
	for _, m := range validMemberships {
		if membership == m {
			return true
		}
	}
	return false
}

func contains(list []string, x string) bool {
	for _, item := range list {
		if item == x {
			return true
		}
	}
	return false
}

func debugLogf(format string, args ...interface{}) {
	if appCfg.debug {
		log.Printf("[DEBUG] "+format, args...)
	}
}

func shallowCloneRequested(shallowList []string, namespace, name string) bool {
	if len(shallowList) == 0 {
		return false
	}
	return contains(shallowList, fmt.Sprintf("%s/%s", namespace, name))
}

func fileExists(filePath string) (bool, error) {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return true, nil
	}
}

func getFileContents(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func writeFileContents(filePath string, content string) error {
	err := os.WriteFile(filePath, []byte(content), 0751)
	if err != nil {
		return err
	}
	return nil
}
