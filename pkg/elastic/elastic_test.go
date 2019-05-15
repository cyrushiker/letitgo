package elastic

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"
)

func TestString(t *testing.T) {
	var d Doc
	t.Logf("json format of a doc:\n%s", &d)

	n := Node{
		ID:   "1",
		GID:  "2",
		CIDs: []string{"2", "3"},
		Kvs: map[string]interface{}{
			"name":     "Cyrushiker",
			"age":      29,
			"birthday": time.Date(1991, time.September, 24, 0, 0, 0, 0, time.UTC),
		},
	}
	t.Logf("Node json:\n%s", &n)
}

func TestOrigin(t *testing.T) {
	var o Origin
	f, err := ioutil.ReadFile("1.json")
	if err != nil {
		t.Fatal(err)
	}
	_ = json.Unmarshal([]byte(f), &o)
	// t.Logf("Origin json:\n%s", &o)
	nodes := o.ParseNodes()
	// for _, n := range nodes {
	// 	t.Logf("%v", n)
	// }
	err = SaveNodes(nodes)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoc(t *testing.T) {
	d := &Doc{
		ID:   1,
		Name: "PCI",
		Groups: []*Group{
			&Group{
				ID:   "1",
				Key:  "patient",
				Name: "患者信息",
				Properties: []*Property{
					NewProperty("2", "name", "姓名", "text"),
					NewProperty("3", "gender", "性别", "keyword"),
					NewProperty("4", "age", "年龄", "long"),
					NewProperty("5", "birthday", "出生日期", "date"),
				},
			},
		},
	}
	t.Logf("Node json:\n%s", d)
	d.CreateRecords()
}
