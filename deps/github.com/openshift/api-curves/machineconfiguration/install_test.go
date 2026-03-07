package machineconfiguration

import (
	"testing"
)

func TestGroupName(t *testing.T) {
	if got, want := GroupName, "machineconfiguration.openshift.io"; got != want {
		t.Fatalf("mismatch group name, got: %s want: %s", got, want)
	}
}
