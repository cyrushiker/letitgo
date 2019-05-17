package emr

import (
	// "reflect"
	"testing"
)

func TestString(t *testing.T) {
	initTables("emdata_emr_172")
	getBatch(projIDs[0], 10, "op")
}

// func TestInfo(t *testing.T) {
// 	// var mi &medicalInfo
// }
