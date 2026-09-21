package ingest

import (
	"strings"
	"testing"
)

func TestVerifyArtifact(t *testing.T) {
	if err := Verify(strings.NewReader("abc"), "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"); err != nil {
		t.Fatal(err)
	}
	if err := Verify(strings.NewReader("changed"), "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"); err == nil {
		t.Fatal("changed artifact accepted")
	}
}
