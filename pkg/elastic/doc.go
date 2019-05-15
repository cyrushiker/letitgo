// The doc struct define

package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

// DefaultDocIndex to store property infos
const DefaultDocIndex = "letitgo-property"

// DefaultDocMapping to define the mapping
const DefaultDocMapping = `
{
	"mappings" : {
		"properties" : {
			"type" : {
				"type" : "keyword"
			},
			"property" : {
				"properties" : {
					"id" : {
						"type" : "keyword"
					},
					"key" : {
						"type" : "keyword"
					},
					"name" : {
						"type" : "keyword"
					},
					"etype" : {
						"type" : "keyword"
					}
				}
			},
			"group" : {
				"properties" : {
					"id" : {
						"type" : "keyword"
					},
					"key" : {
						"type" : "keyword"
					},
					"name" : {
						"type" : "keyword"
					}
				}
			},
			"doc" : {
				"properties" : {
					"id" : {
						"type" : "keyword"
					},
					"code" : {
						"type" : "keyword"
					},
					"name" : {
						"type" : "keyword"
					}
				}
			}
		}
	}
}
`

func init() {
	escli := getEscli()
	res, err := escli.Indices.Exists([]string{DefaultDocIndex})
	if err != nil {
		log.Fatalf("Indices exists %s failed: %s", DefaultDocIndex, err)
	}
	if res.IsError() {
		log.Printf("Index(%s) is not exists then create it with default mapping\n", DefaultDocIndex)
		req := esapi.IndicesCreateRequest{
			Index: DefaultDocIndex,
			Body:  strings.NewReader(DefaultDocMapping),
		}
		res, err := req.Do(context.Background(), escli)
		if err != nil {
			log.Fatalf("Create index(%s) failed: %s", DefaultDocIndex, err)
		}
		if res.IsError() {
			log.Fatalf("Cannot create index(%s): %s", DefaultDocIndex, res)
		}
		log.Printf("Index(%s) is created successfully\n", DefaultDocIndex)
	}
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
	ID         string      `json:"id"`
	Key        string      `json:"key"`
	Name       string      `json:"name"`
	ParentKey  string      `json:"parentKey"`
	Properties []*Property `json:"properties"`
}

// Doc is a set of groups
type Doc struct {
	ID     int32    `json:"id"`
	Name   string   `json:"name"`
	Groups []*Group `json:"groups"`
}

// Record to es
type Record struct {
	Type     string    `json:"type"`
	Doc      *Doc      `json:"doc,omitempty"`
	Group    *Group    `json:"group,omitempty"`
	Property *Property `json:"property,omitempty"`
}

func (d *Doc) String() string {
	ds, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		log.Println(err)
		return "{}"
	}
	return string(ds)
}

// NewProperty new a Property
func NewProperty(id, key, name, etype string) *Property {
	return &Property{ID: id, Key: key, Name: name, Etype: etype}
}

// CreateRecords create records of properties
func (d *Doc) CreateRecords() error {
	escli := getEscli()
	var buf bytes.Buffer
	for _, g := range d.Groups {
		for _, p := range g.Properties {
			r := &Record{Type: "property", Property: p}
			meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "property:%s" } }%s`, p.ID, "\n"))
			data, err := json.Marshal(r)
			// log.Println(string(data))
			if err != nil {
				return fmt.Errorf("Cannot encode node %s: %s", p.ID, err)
			}
			data = append(data, "\n"...)
			buf.Write(meta)
			buf.Write(data)
		}
	}
	log.Println(string(buf.Bytes()))
	res, err := escli.Bulk(bytes.NewReader(buf.Bytes()), escli.Bulk.WithIndex(DefaultDocIndex))
	if err != nil {
		return fmt.Errorf("Failure bulk indexing nodes: %s", err)
	}
	defer res.Body.Close()
	return nil
}
