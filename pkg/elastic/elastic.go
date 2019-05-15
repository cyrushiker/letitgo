// 将一个大的文档按逻辑分解为相互关联的，统一的数据结构，方便存储和查询
// 有限深度的N叉树结构

package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/google/uuid"
)

// DefaultIndex of elastic
const DefaultIndex = "letitgo"

// DefaultMapping for DefaultIndex
const DefaultMapping = `
{
	"mappings" : {
		"properties" : {
			"clen" : {
				"type" : "long"
			},
			"gid" : {
				"type" : "keyword"
			},
			"groupk" : {
				"type" : "keyword"
			},
			"id" : {
				"type" : "keyword"
			}
		}
	}
}
`

var globalEsClient *elasticsearch.Client

func getEscli() *elasticsearch.Client {
	if globalEsClient != nil {
		return globalEsClient
	}
	var err error
	globalEsClient, err = elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Init elastic client failed: %s", err)
	}
	log.Println("Init elastic client to http://localhost:9200")
	return globalEsClient
}

func init() {
	// TODO: 从settings中获取elastic config and init esclient for this pkg
	es := getEscli()
	res, err := es.Indices.Exists([]string{DefaultIndex})
	if err != nil {
		log.Fatalf("Indices exists %s failed: %s", DefaultIndex, err)
	}
	if res.IsError() {
		log.Printf("Index(%s) is not exists then create it with default mapping\n", DefaultIndex)
		req := esapi.IndicesCreateRequest{
			Index: DefaultIndex,
			Body:  strings.NewReader(DefaultMapping),
		}
		res, err := req.Do(context.Background(), es)
		if err != nil {
			log.Fatalf("Create index(%s) failed: %s", DefaultIndex, err)
		}
		if res.IsError() {
			log.Fatalf("Cannot create index(%s): %s", DefaultIndex, res)
		}
		log.Printf("Index(%s) is created successfully\n", DefaultIndex)
	}
}

// Node of a n-tree
type Node struct {
	GID    string                 `json:"gid"`
	ID     string                 `json:"id"`
	CIDs   []string               `json:"cids"`
	Clen   int32                  `json:"clen"`
	Groupk string                 `json:"groupk"`
	Kvs    map[string]interface{} `json:"kvs"`
}

func (n *Node) String() string {
	s, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		log.Println(err)
		return "{}"
	}
	return string(s)
}

// SaveNodes save nodes to elastic
func SaveNodes(nodes []*Node) error {
	escli := getEscli()
	var buf bytes.Buffer
	for _, n := range nodes {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, n.ID, "\n"))
		data, err := json.Marshal(n)
		if err != nil {
			return fmt.Errorf("Cannot encode node %s: %s", n.ID, err)
		}
		data = append(data, "\n"...)
		buf.Write(meta)
		buf.Write(data)
	}
	res, err := escli.Bulk(bytes.NewReader(buf.Bytes()), escli.Bulk.WithIndex(DefaultIndex))
	if err != nil {
		return fmt.Errorf("Failure bulk indexing nodes: %s", err)
	}
	defer res.Body.Close()
	return nil
}

// Origin disease info
type Origin struct {
	DiseaseID  int                                 `json:"diseaseId"`
	HospitalID int                                 `json:"hospitalId"`
	DeptID     int                                 `json:"deptId"`
	SourceCode string                              `json:"sourceCode"`
	SourceID   string                              `json:"sourceId"`
	Docs       map[string][]map[string]interface{} `json:"docs"`
}

func (o *Origin) String() string {
	s, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		log.Println(err)
		return "{}"
	}
	return string(s)
}

// ParseNodes parse the Origin to nodes
func (o *Origin) ParseNodes() []*Node {
	var nodes []*Node
	guid, _ := uuid.NewRandom()
	gid := guid.String()
	for k, v := range o.Docs {
		for _, d := range v {
			uid, _ := uuid.NewUUID()
			n := Node{GID: gid, ID: uid.String(), Kvs: d, Groupk: k}
			nodes = append(nodes, &n)
		}
	}
	return nodes
}
