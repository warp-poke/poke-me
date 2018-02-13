package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

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
	auth.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	return &Cloner{
		sshTransportClient: auth,
		gitRepo:            gitRepo,
		path:               path,
	}, nil
}

// Clone replace existing content by master
func (c *Cloner) Clone(sha string, backup bool) error {

	if backup {
		// Backup old scripts
		backupFolder := fmt.Sprintf("%s.%s", c.path, time.Now())
		if err := os.Rename(c.path, backupFolder); err != nil {
			return fmt.Errorf("Failed to backup current scripts: %s", err.Error())
		}
	}

	if err := os.RemoveAll(c.path); err != nil {
		return fmt.Errorf("Failed to remove current scripts directory: %s", err.Error())
	}

	repo, err := git.PlainClone(c.path, false, &git.CloneOptions{
		URL:           c.gitRepo,
		Auth:          c.sshTransportClient,
		ReferenceName: plumbing.Master,
		Depth:         1,
		SingleBranch:  true,
	})
	if err != nil {
		return fmt.Errorf("Failed to clone repo: %s", err.Error())
	}

	tree, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = tree.Checkout(&git.CheckoutOptions{
		Hash:  plumbing.NewHash(sha),
		Force: true,
	})
	if err != nil {
		return err
	}

	return nil
}
