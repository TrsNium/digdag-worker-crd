// +build !ignore_autogenerated

/*

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

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HorizontalDigdagWorkerAutoscaler) DeepCopyInto(out *HorizontalDigdagWorkerAutoscaler) {
	*out = *in
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HorizontalDigdagWorkerAutoscaler.
func (in *HorizontalDigdagWorkerAutoscaler) DeepCopy() *HorizontalDigdagWorkerAutoscaler {
	if in == nil {
		return nil
	}
	out := new(HorizontalDigdagWorkerAutoscaler)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HorizontalDigdagWorkerAutoscaler) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HorizontalDigdagWorkerAutoscalerList) DeepCopyInto(out *HorizontalDigdagWorkerAutoscalerList) {
	*out = *in
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]HorizontalDigdagWorkerAutoscaler, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HorizontalDigdagWorkerAutoscalerList.
func (in *HorizontalDigdagWorkerAutoscalerList) DeepCopy() *HorizontalDigdagWorkerAutoscalerList {
	if in == nil {
		return nil
	}
	out := new(HorizontalDigdagWorkerAutoscalerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HorizontalDigdagWorkerAutoscalerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HorizontalDigdagWorkerAutoscalerSpec) DeepCopyInto(out *HorizontalDigdagWorkerAutoscalerSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HorizontalDigdagWorkerAutoscalerSpec.
func (in *HorizontalDigdagWorkerAutoscalerSpec) DeepCopy() *HorizontalDigdagWorkerAutoscalerSpec {
	if in == nil {
		return nil
	}
	out := new(HorizontalDigdagWorkerAutoscalerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HorizontalDigdagWorkerAutoscalerStatus) DeepCopyInto(out *HorizontalDigdagWorkerAutoscalerStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HorizontalDigdagWorkerAutoscalerStatus.
func (in *HorizontalDigdagWorkerAutoscalerStatus) DeepCopy() *HorizontalDigdagWorkerAutoscalerStatus {
	if in == nil {
		return nil
	}
	out := new(HorizontalDigdagWorkerAutoscalerStatus)
	in.DeepCopyInto(out)
	return out
}
