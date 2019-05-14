// 将一个大的文档按逻辑分解为相互关联的，统一的数据结构，方便存储和查询
// 有限深度的N叉树结构

package elastic

import (
	"encoding/json"
	"log"
)

// Bdoc is the root for a n-tree
type Bdoc struct {
	DocID string
	ID    string
	Value map[string]interface{}
}

// Sdoc is the leaf node for the tree
type Sdoc struct {
	DocID string
	ID    string
	PID   string
	Value map[string]interface{}
}

// Property describe the minimum element of a doc
type Property struct {
	ID    string `json:"id"`
	Key   string `json:"key"`
	Name  string `json:"name"`
	Etype string `json:"etype"`
}

// Group some property together
type Group struct {
	ID         string     `json:"id"`
	Key        string     `json:"key"`
	Name       string     `json:"name"`
	ParentKey  string     `json:"parentKey"`
	Properties []Property `json:"properties"`
}

// Disease is a set of groups
type Disease struct {
	ID     int32   `json:"id"`
	Name   string  `json:"name"`
	Groups []Group `json:"groups"`
}

func (d *Disease) String() string {
	ds, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		log.Println(err)
		return "{}"
	}
	return string(ds)
}
