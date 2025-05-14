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

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FileContentEncoding indicates the type of file encoding.
type FileContentEncoding string

// List of file encodings
const (
	FileContentEncodingText   FileContentEncoding = "text"
	FileContentEncodingBase64 FileContentEncoding = "base64"
)

// FileParameters define the desired state of a Gitlab repository file
// https://docs.gitlab.com/api/repository_files/
type FileParameters struct {
	// ProjectID is the ID of the project to create the file in.
	// +optional
	// +immutable
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-gitlab/apis/projects/v1alpha1.Project
	ProjectID *string `json:"projectId,omitempty"`

	// ProjectIDRef is a reference to a project to retrieve its projectId.
	// +optional
	// +immutable
	ProjectIDRef *xpv1.Reference `json:"projectIdRef,omitempty"`

	// ProjectIDSelector selects reference to a project to retrieve its projectId.
	// +optional
	ProjectIDSelector *xpv1.Selector `json:"projectIdSelector,omitempty"`

	// Name of the new branch to create. The commit is added to this branch.
	// +required
	Branch *string `json:"branch"`

	// The commit message.
	// +required
	CommitMessage *string `json:"commitMessage"`

	// The file’s content.
	// +required
	Content *string `json:"content"`

	// Regex to check URL-encoding pattern!
	// URL-encoded full path to new file. For example: lib%2Fclass%2Erb.
	// +required
	FilePath *string `json:"filePath"`

	// The commit author’s email address.
	// +optional
	AuthorEmail *string `json:"authorEmail,omitempty"`

	// The commit author’s name.
	// +optional
	AuthorName *string `json:"authorName,omitempty"`

	// Change encoding to base64. Default is text.
	// +kubebuilder:validation:Enum:=text;base64
	// +optional
	Encoding *string `json:"encoding,omitempty"`

	// Enables or disables the execute flag on the file. Can be true or false.
	// +optional
	ExecuteFilemode *bool `json:"executeFilemode,omitempty"`

	// Name of the base branch to create the new branch from.
	// +optional
	StartBranch *string `json:"startBranch,omitempty"`
}

type FileObservation struct {
	FileName        string `json:"fileName"`
	FilePath        string `json:"filePath"`
	Size            int    `json:"size"`
	Encoding        string `json:"encoding"`
	Content         string `json:"content"`
	ContentSHA256   string `json:"ContentSHA256"`
	Ref             string `json:"ref"`
	BlobID          string `json:"blobID"`
	CommitID        string `json:"commitID"`
	LastCommitID    string `json:"lastCommitID"`
	ExecuteFilemode bool   `json:"executeFilemode"`
}

// A FileSpec defines the desired state of a Gitlab repository file.
type FileSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       FileParameters `json:"forProvider"`
}

// A FileStatus represents the observed state of a Gitlab repository file.
type FileStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          FileObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A File is a managed resource that represents a Gitlab repository file.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gitlab}
type File struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FileSpec   `json:"spec"`
	Status FileStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FileList contains a list of File items.
type FileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []File `json:"items"`
}
