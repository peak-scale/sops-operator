//go:build !ignore_autogenerated

/*
Copyright 2025.

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

package api

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Agekey) DeepCopyInto(out *Agekey) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Agekey.
func (in *Agekey) DeepCopy() *Agekey {
	if in == nil {
		return nil
	}
	out := new(Agekey)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Azkvkey) DeepCopyInto(out *Azkvkey) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Azkvkey.
func (in *Azkvkey) DeepCopy() *Azkvkey {
	if in == nil {
		return nil
	}
	out := new(Azkvkey)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GcpKmskey) DeepCopyInto(out *GcpKmskey) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GcpKmskey.
func (in *GcpKmskey) DeepCopy() *GcpKmskey {
	if in == nil {
		return nil
	}
	out := new(GcpKmskey)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Keygroup) DeepCopyInto(out *Keygroup) {
	*out = *in
	if in.Pgpkeys != nil {
		in, out := &in.Pgpkeys, &out.Pgpkeys
		*out = make([]Pgpkey, len(*in))
		copy(*out, *in)
	}
	if in.Kmskeys != nil {
		in, out := &in.Kmskeys, &out.Kmskeys
		*out = make([]Kmskey, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.GcpKmskeys != nil {
		in, out := &in.GcpKmskeys, &out.GcpKmskeys
		*out = make([]GcpKmskey, len(*in))
		copy(*out, *in)
	}
	if in.AzureKeyVaultkeys != nil {
		in, out := &in.AzureKeyVaultkeys, &out.AzureKeyVaultkeys
		*out = make([]Azkvkey, len(*in))
		copy(*out, *in)
	}
	if in.Vaultkeys != nil {
		in, out := &in.Vaultkeys, &out.Vaultkeys
		*out = make([]Vaultkey, len(*in))
		copy(*out, *in)
	}
	if in.Agekeys != nil {
		in, out := &in.Agekeys, &out.Agekeys
		*out = make([]Agekey, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Keygroup.
func (in *Keygroup) DeepCopy() *Keygroup {
	if in == nil {
		return nil
	}
	out := new(Keygroup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Kmskey) DeepCopyInto(out *Kmskey) {
	*out = *in
	if in.Context != nil {
		in, out := &in.Context, &out.Context
		*out = make(map[string]*string, len(*in))
		for key, val := range *in {
			var outVal *string
			if val == nil {
				(*out)[key] = nil
			} else {
				inVal := (*in)[key]
				in, out := &inVal, &outVal
				*out = new(string)
				**out = **in
			}
			(*out)[key] = outVal
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Kmskey.
func (in *Kmskey) DeepCopy() *Kmskey {
	if in == nil {
		return nil
	}
	out := new(Kmskey)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Metadata) DeepCopyInto(out *Metadata) {
	*out = *in
	if in.KeyGroups != nil {
		in, out := &in.KeyGroups, &out.KeyGroups
		*out = make([]Keygroup, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Kmskeys != nil {
		in, out := &in.Kmskeys, &out.Kmskeys
		*out = make([]Kmskey, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.GcpKmskeys != nil {
		in, out := &in.GcpKmskeys, &out.GcpKmskeys
		*out = make([]GcpKmskey, len(*in))
		copy(*out, *in)
	}
	if in.AzureKeyVaultkeys != nil {
		in, out := &in.AzureKeyVaultkeys, &out.AzureKeyVaultkeys
		*out = make([]Azkvkey, len(*in))
		copy(*out, *in)
	}
	if in.Vaultkeys != nil {
		in, out := &in.Vaultkeys, &out.Vaultkeys
		*out = make([]Vaultkey, len(*in))
		copy(*out, *in)
	}
	if in.Agekeys != nil {
		in, out := &in.Agekeys, &out.Agekeys
		*out = make([]Agekey, len(*in))
		copy(*out, *in)
	}
	if in.Pgpkeys != nil {
		in, out := &in.Pgpkeys, &out.Pgpkeys
		*out = make([]Pgpkey, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Metadata.
func (in *Metadata) DeepCopy() *Metadata {
	if in == nil {
		return nil
	}
	out := new(Metadata)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NamespacedSelector) DeepCopyInto(out *NamespacedSelector) {
	*out = *in
	if in.LabelSelector != nil {
		in, out := &in.LabelSelector, &out.LabelSelector
		*out = new(v1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	if in.NamespaceSelector != nil {
		in, out := &in.NamespaceSelector, &out.NamespaceSelector
		*out = new(v1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NamespacedSelector.
func (in *NamespacedSelector) DeepCopy() *NamespacedSelector {
	if in == nil {
		return nil
	}
	out := new(NamespacedSelector)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Pgpkey) DeepCopyInto(out *Pgpkey) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Pgpkey.
func (in *Pgpkey) DeepCopy() *Pgpkey {
	if in == nil {
		return nil
	}
	out := new(Pgpkey)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Vaultkey) DeepCopyInto(out *Vaultkey) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Vaultkey.
func (in *Vaultkey) DeepCopy() *Vaultkey {
	if in == nil {
		return nil
	}
	out := new(Vaultkey)
	in.DeepCopyInto(out)
	return out
}
