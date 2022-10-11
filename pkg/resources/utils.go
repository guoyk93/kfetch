package resources

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"strings"
)

var (
	int32Zero = int32(0)
)

func HasPrefix(name string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func JSON2YAML(buf []byte) (out []byte, err error) {
	var m map[string]interface{}
	if err = json.Unmarshal(buf, &m); err != nil {
		return
	}
	out, err = yaml.Marshal(m)
	return
}

func YAML2JSON(buf []byte) (out []byte, err error) {
	var m map[string]interface{}
	if err = yaml.Unmarshal(buf, &m); err != nil {
		return
	}
	out, err = json.Marshal(m)
	return
}
