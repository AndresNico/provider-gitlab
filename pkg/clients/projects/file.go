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

	gitlab "gitlab.com/gitlab-org/api/client-go"
	"golang.org/x/net/context"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1"
	"github.com/crossplane-contrib/provider-gitlab/pkg/clients"
)

const (
	errFileNotFound = "404 Not found"
)

// FileClient defines Gitlab file service operations
type FileClient interface {
	GetFile(pid interface{}, fileName string, opt *gitlab.GetFileOptions, options ...gitlab.RequestOptionFunc) (*gitlab.File, *gitlab.Response, error)
	CreateFile(pid interface{}, fileName string, opt *gitlab.CreateFileOptions, options ...gitlab.RequestOptionFunc) (*gitlab.FileInfo, *gitlab.Response, error)
	UpdateFile(pid interface{}, fileName string, opt *gitlab.UpdateFileOptions, options ...gitlab.RequestOptionFunc) (*gitlab.FileInfo, *gitlab.Response, error)
	DeleteFile(pid interface{}, fileName string, opt *gitlab.DeleteFileOptions, options ...gitlab.RequestOptionFunc) (*gitlab.Response, error)
}

// NewFileClient returns a new Gitlab file service
func NewFileClient(cfg clients.Config) FileClient {
	git := clients.NewClient(cfg)
	return git.RepositoryFiles
}

// IsFileNotFound helper function to test for errFileNotFound error.
func IsFileNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errFileNotFound)
}

func LateInitializeFile(in *v1alpha1.FileParameters, file *gitlab.File) {
	if file == nil {
		return
	}

	if in.Encoding == nil {
		in.Encoding = &file.Encoding
	}

	if in.ExecuteFilemode == nil {
		in.ExecuteFilemode = &file.ExecuteFilemode
	}
}

func IsFileUpToDate(p *v1alpha1.FileParameters, g *gitlab.File) bool {
	return true
}

// GenerateGetFileOptions generates file get options
func GenerateGetFileOptions(p *v1alpha1.FileParameters, client client.Client, ctx context.Context) *gitlab.GetFileOptions {
	return &gitlab.GetFileOptions{Ref: p.Branch}
}

// GenerateFileObservation is used to produce v1alpha1.FileObservation from
// gitlab.File.
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
		SHA256:          file.SHA256,
		Ref:             file.Ref,
		BlobID:          file.BlobID,
		CommitID:        file.CommitID,
		LastCommitID:    file.LastCommitID,
		ExecuteFilemode: file.ExecuteFilemode,
	}

	return o
}

// GenerateCreateFileOptions generates file creation options
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

// GenerateEditFileOptions generates file edit options
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

// GenerateDeleteFileOptions generates file delete options
func GenerateDeleteFileOptions(p *v1alpha1.FileParameters, client client.Client, ctx context.Context) *gitlab.DeleteFileOptions {
	o := &gitlab.DeleteFileOptions{
		Branch:        p.Branch,
		StartBranch:   p.StartBranch,
		AuthorEmail:   p.AuthorEmail,
		AuthorName:    p.AuthorName,
		CommitMessage: p.CommitMessage,
		// LastCommitID: p.LastCommitID,
	}

	return o
}
