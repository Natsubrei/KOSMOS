package hostid

import "testing"

func TestResolveOverrideWins(t *testing.T) {
	got, err := Resolve("web-1")
	if err != nil {
		t.Fatal(err)
	}
	if got != "web-1" {
		t.Fatalf("got %q, want web-1", got)
	}
}
