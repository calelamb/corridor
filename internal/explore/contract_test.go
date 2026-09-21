package explore

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestExplorationContract(t *testing.T) {
	l := openapi3.NewLoader()
	d, err := l.LoadFromFile("../../api/exploration.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err = d.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if d.Paths.Len() != 5 {
		t.Fatal("missing public contract routes")
	}
}
