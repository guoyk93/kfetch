package resources

import (
	"encoding/json"
	jsonpatch "github.com/evanphx/json-patch"
)

const (
	OpAdd     = "add"
	OpRemove  = "remove"
	OpReplace = "replace"
	OpCopy    = "copy"
	OpMove    = "move"
	OpTest    = "test"
)

type Patch struct {
	Op    string      `json:"op,omitempty"`
	Path  string      `json:"path,omitempty"`
	From  string      `json:"from,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

type Patches []Patch

type PatchSet []Patches

func (ps PatchSet) Apply(data []byte) (out []byte, err error) {
	for _, item := range ps {
		var buf []byte
		if buf, err = json.Marshal(item); err != nil {
			return
		}
		var patch jsonpatch.Patch
		if patch, err = jsonpatch.DecodePatch(buf); err != nil {
			return
		}
		if buf, err = patch.Apply(data); err == nil {
			data = buf
		}
	}
	out = data
	err = nil
	return
}

var (
	defaultSanitizers = PatchSet{
		{{Op: OpRemove, Path: "/status"}},
		{{Op: OpRemove, Path: "/metadata/namespace"}},
		{{Op: OpRemove, Path: "/metadata/creationTimestamp"}},
		{{Op: OpRemove, Path: "/metadata/generation"}},
		{{Op: OpRemove, Path: "/metadata/selfLink"}},
		{{Op: OpRemove, Path: "/metadata/uid"}},
		{{Op: OpRemove, Path: "/metadata/managedFields"}},
		{{Op: OpRemove, Path: "/metadata/finalizers"}},
		{{Op: OpRemove, Path: "/metadata/annotations/kubectl.kubernetes.io~1last-applied-configuration"}},
		{{Op: OpRemove, Path: "/metadata/annotations/deployment.kubernetes.io~1revision"}},
		{{Op: OpRemove, Path: "/metadata/annotations/field.cattle.io~1ingressState"}},
		{{Op: OpRemove, Path: "/metadata/annotations/field.cattle.io~1publicEndpoints"}},
		{{Op: OpRemove, Path: "/spec/template/metadata/creationTimestamp"}},
		{{Op: OpRemove, Path: "/spec/replicas"}},
	}
	noResourceVersionSanitizers = PatchSet{
		{{Op: OpRemove, Path: "/metadata/resourceVersion"}},
	}
)
