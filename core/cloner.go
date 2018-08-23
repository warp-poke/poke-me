package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/cbroglie/mustache"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	sshTransport "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

// Cloner can clone a Git repo into a path, with backup
type Cloner struct {
	sshTransportClient *sshTransport.PublicKeys
	gitRepo            string
	path               string
}

// NewCloner instantiate a couple git Repo => path
// PATH = runner.root (Warp10)
func NewCloner(sshTransportKey, gitRepo, path string) (*Cloner, error) {

	privKeyBytes, err := ioutil.ReadFile(sshTransportKey)
	if err != nil {
		return nil, err
	}

	auth, err := sshTransport.NewPublicKeys("git", privKeyBytes, "")
	if err != nil {
		return nil, err
	}
	auth.HostKeyCallbackHelper.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	return &Cloner{
		sshTransportClient: auth,
		gitRepo:            gitRepo,
		path:               strings.TrimSuffix(path, "/"),
	}, nil
}

// Clone replace existing content by master
func (c *Cloner) Clone(sha string, secrets map[string]string, backup bool) error {

	if backup {
		// Backup old scripts
		backupFolder := fmt.Sprintf("%s.%s", c.path, time.Now())
		if _, err := os.Stat(c.path); err == nil {
			if err := os.Rename(c.path, backupFolder); err != nil {
				log.Error(err)
				return fmt.Errorf("Failed to backup current scripts: %s", err.Error())
			}
		}
	}

	if _, err := os.Stat(c.path); err == nil {
		if err := os.RemoveAll(c.path); err != nil {
			log.Error(err)
			return fmt.Errorf("Failed to remove current scripts directory: %s", err.Error())
		}
	}

	if err := os.Mkdir(c.path, os.ModePerm); err != nil {
		log.WithError(err).Warn("Failed to create ws dir")
	}

	repo, err := git.PlainClone(c.path, false, &git.CloneOptions{
		URL:   c.gitRepo,
		Auth:  c.sshTransportClient,
		Depth: 1,
	})
	if err != nil {
		log.Error(err)
		return fmt.Errorf("Failed to clone repo: %s", err.Error())
	}

	tree, err := repo.Worktree()
	if err != nil {
		log.WithError(err).Error("Failed to get worktree")
		return err
	}

	err = tree.Checkout(&git.CheckoutOptions{
		Hash:  plumbing.NewHash(sha),
		Force: true,
	})
	if err != nil {
		log.WithError(err).Error("Failed to checkout")
		return err
	}

	dirs, err := tree.Filesystem.ReadDir(".")
	if err != nil {
		log.WithError(err).Error("Failed to read dir")
		return err
	}

	for k := range secrets {
		log.Debug("known secret", k)
	}

	// Replace secrets
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		if dir.Name() == ".git" {
			continue
		}
		log.Debug("Scan dir %s", dir.Name())

		files, err := tree.Filesystem.ReadDir(dir.Name())
		if err != nil {
			log.WithError(err).Error("Failed to read file")
			return err
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}
			log.Debug("Scan file %s", file.Name())

			absFilePath := path.Join(c.path, dir.Name(), file.Name())

			tplFile, err := mustache.RenderFile(absFilePath, secrets)
			if err != nil {
				log.WithError(err).Error("Failed to render file")
				return err
			}
			log.Debug(tplFile)

			if err := ioutil.WriteFile(absFilePath, []byte(tplFile), os.ModePerm); err != nil {
				log.WithError(err).Error("Failed to write file")
				return err
			}
		}
	}

	return nil
}
