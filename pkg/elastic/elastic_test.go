package elastic

import (
	"testing"
)

func TestString(t *testing.T) {
	var d Disease
	t.Logf("json format of disease:\n%s", &d)
}
