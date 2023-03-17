//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2022.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommitTemplate) DeepCopyInto(out *CommitTemplate) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommitTemplate.
func (in *CommitTemplate) DeepCopy() *CommitTemplate {
	if in == nil {
		return nil
	}
	out := new(CommitTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitSync) DeepCopyInto(out *GitSync) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitSync.
func (in *GitSync) DeepCopy() *GitSync {
	if in == nil {
		return nil
	}
	out := new(GitSync)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GitSync) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitSyncList) DeepCopyInto(out *GitSyncList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GitSync, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitSyncList.
func (in *GitSyncList) DeepCopy() *GitSyncList {
	if in == nil {
		return nil
	}
	out := new(GitSyncList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GitSyncList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitSyncSpec) DeepCopyInto(out *GitSyncSpec) {
	*out = *in
	out.SnapshotRef = in.SnapshotRef
	out.Interval = in.Interval
	out.AuthRef = in.AuthRef
	if in.CommitTemplate != nil {
		in, out := &in.CommitTemplate, &out.CommitTemplate
		*out = new(CommitTemplate)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitSyncSpec.
func (in *GitSyncSpec) DeepCopy() *GitSyncSpec {
	if in == nil {
		return nil
	}
	out := new(GitSyncSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitSyncStatus) DeepCopyInto(out *GitSyncStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitSyncStatus.
func (in *GitSyncStatus) DeepCopy() *GitSyncStatus {
	if in == nil {
		return nil
	}
	out := new(GitSyncStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Ref) DeepCopyInto(out *Ref) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Ref.
func (in *Ref) DeepCopy() *Ref {
	if in == nil {
		return nil
	}
	out := new(Ref)
	in.DeepCopyInto(out)
	return out
}
