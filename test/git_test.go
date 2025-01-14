package test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"testing"

	vcs "github.com/Buliqioqiolibusdo/demp-vcs"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/require"
)

func TestNewGitClient_Existing(t *testing.T) {
	var err error
	T.Setup(t)

	// git client
	c, err := vcs.NewGitClient(
		vcs.WithPath(T.LocalRepo.GetPath()),
	)
	require.Nil(t, err)
	require.NotEmpty(t, c.GetRepository())
}

func TestNewGitClient_Fs(t *testing.T) {
	var err error
	T.Setup(t)

	// git client
	c, err := vcs.NewGitClient(
		vcs.WithPath(T.FsRepoPath),
		vcs.WithRemoteUrl(T.RemoteRepoPath),
	)
	require.Nil(t, err)
	require.NotEmpty(t, c.GetRepository())
	require.Equal(t, T.RemoteRepoPath, c.GetRemoteUrl())
}

func TestNewGitClient_Mem(t *testing.T) {
	var err error
	T.Setup(t)

	// git client
	c, err := vcs.NewGitClient(
		vcs.WithPath(T.MemRepoPath),
		vcs.WithRemoteUrl(T.RemoteRepoPath),
		vcs.WithIsMem(),
	)
	require.Nil(t, err)
	require.NotEmpty(t, c.GetRepository())
}

func TestGitClient_CommitAllAndCheckoutBranch(t *testing.T) {
	var err error
	T.Setup(t)

	// commit
	filePath := path.Join(T.LocalRepoPath, T.TestFileName)
	err = ioutil.WriteFile(filePath, []byte(T.TestFileContent), os.FileMode(0766))
	require.Nil(t, err)
	err = T.LocalRepo.CommitAll(T.TestCommitMessage)
	require.Nil(t, err)

	// checkout branch
	err = T.LocalRepo.CheckoutBranch(T.TestBranchName)
	require.Nil(t, err)

	// validate
	branch, err := T.LocalRepo.GetCurrentBranch()
	require.Nil(t, err)
	require.Equal(t, plumbing.NewBranchReferenceName(T.TestBranchName).String(), branch)
}

func TestGitClient_Push(t *testing.T) {
	var err error
	T.Setup(t)

	// commit
	filePath := path.Join(T.LocalRepoPath, T.TestFileName)
	err = ioutil.WriteFile(filePath, []byte(T.TestFileContent), os.FileMode(0766))
	require.Nil(t, err)
	err = T.LocalRepo.CommitAll(T.TestCommitMessage)
	require.Nil(t, err)

	// push
	err = T.LocalRepo.Push()
	require.Nil(t, err)
}

func TestGitClient_Reset(t *testing.T) {
	var err error
	T.Setup(t)

	// file
	filePath := path.Join(T.LocalRepoPath, T.TestFileName)
	err = ioutil.WriteFile(filePath, []byte(T.TestFileContent), os.FileMode(0766))
	require.Nil(t, err)

	// reset
	err = T.LocalRepo.Reset(vcs.WithMode(git.HardReset)) // git reset --hard
	require.Nil(t, err)
	_, err = os.Stat(filePath)
	require.IsType(t, &os.PathError{}, err)
}

func TestGitClient_GetLogs(t *testing.T) {
	var err error
	T.Setup(t)

	// file
	filePath := path.Join(T.LocalRepoPath, T.TestFileName)
	err = ioutil.WriteFile(filePath, []byte(T.TestFileContent), os.FileMode(0766))
	require.Nil(t, err)
	err = T.LocalRepo.CommitAll(T.TestCommitMessage)
	require.Nil(t, err)

	// get logs
	logs, err := T.LocalRepo.GetLogs()
	require.Nil(t, err)
	require.Greater(t, len(logs), 0)
	require.Equal(t, T.TestCommitMessage, logs[0].Msg)
}

func TestGitClient_InitWithHttpAuth(t *testing.T) {
	var err error
	T.Setup(t)

	// get credentials
	var cred Credential
	data, err := ioutil.ReadFile("credentials.json")
	require.Nil(t, err)
	err = json.Unmarshal(data, &cred)
	require.Nil(t, err)

	// create new git client
	c, err := vcs.NewGitClient(
		vcs.WithPath(T.AuthRepoPath),
		vcs.WithRemoteUrl(cred.TestRepoHttpUrl),
		vcs.WithAuthType(vcs.GitAuthTypeHTTP),
		vcs.WithUsername(cred.Username),
		vcs.WithPassword(cred.Password),
	)
	require.Nil(t, err)
	require.Equal(t, cred.TestRepoHttpUrl, c.GetRemoteUrl())
	require.Equal(t, vcs.GitAuthTypeHTTP, c.GetAuthType())
	require.Equal(t, cred.Username, c.GetUsername())

	// validate
	files, err := ioutil.ReadDir(T.AuthRepoPath)
	require.Greater(t, len(files), 0)
	data, err = ioutil.ReadFile(path.Join(T.AuthRepoPath, "README.md"))
	require.Nil(t, err)
}

func TestGitClient_InitWithSshAuth_PrivateKey(t *testing.T) {
	var err error
	T.Setup(t)

	// get credentials
	var cred Credential
	data, err := ioutil.ReadFile("credentials.json")
	require.Nil(t, err)
	err = json.Unmarshal(data, &cred)
	require.Nil(t, err)

	// git client
	c, err := vcs.NewGitClient(
		vcs.WithPath(T.AuthRepoPath),
		vcs.WithRemoteUrl(cred.TestRepoSshUrl),
		vcs.WithAuthType(vcs.GitAuthTypeSSH),
		vcs.WithUsername(cred.Username),
		vcs.WithPassword(cred.Password),
		vcs.WithPrivateKey(cred.PrivateKey),
	)
	require.Nil(t, err)
	require.Equal(t, cred.TestRepoSshUrl, c.GetRemoteUrl())
	require.Equal(t, vcs.GitAuthTypeSSH, c.GetAuthType())
	require.Equal(t, cred.Username, c.GetUsername())

	// validate
	files, err := ioutil.ReadDir(T.AuthRepoPath)
	require.Greater(t, len(files), 0)
	data, err = ioutil.ReadFile(path.Join(T.AuthRepoPath, "README.md"))
	require.Nil(t, err)
}

func TestGitClient_InitWithSshAuth_PrivateKeyPath(t *testing.T) {
	var err error
	T.Setup(t)

	// get credentials
	var cred Credential
	data, err := ioutil.ReadFile("credentials.json")
	require.Nil(t, err)
	err = json.Unmarshal(data, &cred)
	require.Nil(t, err)

	// git client
	c, err := vcs.NewGitClient(
		vcs.WithPath(T.AuthRepoPath),
		vcs.WithRemoteUrl(cred.TestRepoSshUrl),
		vcs.WithAuthType(vcs.GitAuthTypeSSH),
		vcs.WithUsername(cred.Username),
		vcs.WithPassword(cred.Password),
		vcs.WithPrivateKeyPath(cred.PrivateKeyPath),
	)
	require.Nil(t, err)
	require.Equal(t, cred.TestRepoSshUrl, c.GetRemoteUrl())
	require.Equal(t, vcs.GitAuthTypeSSH, c.GetAuthType())
	require.Equal(t, cred.Username, c.GetUsername())
	require.Equal(t, cred.PrivateKeyPath, c.GetPrivateKeyPath())

	// validate
	files, err := ioutil.ReadDir(T.AuthRepoPath)
	require.Greater(t, len(files), 0)
	data, err = ioutil.ReadFile(path.Join(T.AuthRepoPath, "README.md"))
	require.Nil(t, err)
}

func TestGitClient_Dispose_Fs(t *testing.T) {
	var err error
	T.Setup(t)

	// git client
	c, err := vcs.NewGitClient(
		vcs.WithPath(T.FsRepoPath),
		vcs.WithRemoteUrl(T.RemoteRepoPath),
	)
	require.Nil(t, err)

	// path exists
	require.DirExists(t, T.FsRepoPath)

	// dispose
	err = c.Dispose()
	require.Nil(t, err)

	// validate
	_, err = os.Stat(T.FsRepoPath)
	require.IsType(t, &os.PathError{}, err)
}

func TestGitClient_Dispose_Mem(t *testing.T) {
	var err error
	T.Setup(t)

	// git client
	c, err := vcs.NewGitClient(
		vcs.WithPath(T.MemRepoPath),
		vcs.WithRemoteUrl(T.RemoteRepoPath),
		vcs.WithIsMem(),
	)
	require.Nil(t, err)

	// mem map exists
	stItem, ok := vcs.GitMemStorages.Load(T.MemRepoPath)
	require.True(t, ok)
	require.IsType(t, &memory.Storage{}, stItem)
	fsItem, ok := vcs.GitMemFileSystem.Load(T.MemRepoPath)
	require.True(t, ok)
	require.IsType(t, memfs.New(), fsItem)

	// dispose
	err = c.Dispose()
	require.Nil(t, err)

	// validate
	_, ok = vcs.GitMemStorages.Load("./tmp/test_repo")
	require.False(t, ok)
	_, ok = vcs.GitMemFileSystem.Load("./tmp/test_repo")
	require.False(t, ok)
}
