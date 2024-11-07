/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package projects

import (
	"strings"

	"github.com/xanzy/go-gitlab"
	"golang.org/x/net/context"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
)

const (
	errFileNotFound = "404 Not found"
)

// HookClient defines Gitlab Hook service operations
type FileClient interface {
	GetFile(pid interface{}, file string, opt *gitlab.GetFileOptions, options ...gitlab.RequestOptionFunc) (*gitlab.File, *gitlab.Response, error)
	CreateFile(pid interface{}, file string, opt *gitlab.CreateFileOptions, options ...gitlab.RequestOptionFunc) (*gitlab.FileInfo, *gitlab.Response, error)
	UpdateFile(pid interface{}, file string, opt *gitlab.UpdateFileOptions, options ...gitlab.RequestOptionFunc) (*gitlab.FileInfo, *gitlab.Response, error)
	DeleteFile(pid interface{}, file string, opt *gitlab.DeleteFileOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewHookClient returns a new Gitlab Project service
func NewFileClient(cfg clients.Config) FileClient {
	git := clients.NewClient(cfg)
	return git.RepositoryFiles
}

// IsErrorHookNotFound helper function to test for errProjectNotFound error.
func IsFileNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errFileNotFound)
}

// GenerateHookObservation is used to produce v1alpha1.HookObservation from
// gitlab.Hook.
func GenerateFileObservation(file *gitlab.File) v1alpha1.FileObservation {
	if file == nil {
		return v1alpha1.FileObservation{}
	}

	o := v1alpha1.FileObservation{
		FileName:        file.FileName,
		FilePath:        file.FilePath,
		Size:            file.Size,
		Encoding:        file.Encoding,
		Content:         file.Content,
		ExecuteFilemode: file.ExecuteFilemode,
		Ref:             file.Ref,
		BlobID:          file.BlobID,
		CommitID:        file.CommitID,
		SHA256:          file.SHA256,
		LastCommitID:    file.LastCommitID,
	}

	return o
}

// GenerateCreateHookOptions generates project creation options
func GenerateCreateFileOptions(p *v1alpha1.FileParameters, client client.Client, ctx context.Context) *gitlab.CreateFileOptions {

	file := &gitlab.CreateFileOptions{
		Branch:          p.Branch,
		StartBranch:     p.StartBranch,
		Encoding:        p.Encoding,
		AuthorEmail:     p.AuthorEmail,
		AuthorName:      p.AuthorName,
		Content:         p.Content,
		CommitMessage:   p.CommitMessage,
		ExecuteFilemode: p.ExecuteFilemode,
	}

	return file
}

// GenerateEditHookOptions generates project edit options
func GenerateUpdateFileOptions(p *v1alpha1.FileParameters, client client.Client, ctx context.Context) *gitlab.UpdateFileOptions {

	o := &gitlab.UpdateFileOptions{
		Branch:          p.Branch,
		StartBranch:     p.StartBranch,
		Encoding:        p.Encoding,
		AuthorEmail:     p.AuthorEmail,
		AuthorName:      p.AuthorName,
		Content:         p.Content,
		CommitMessage:   p.CommitMessage,
		ExecuteFilemode: p.ExecuteFilemode,
	}

	return o
}
