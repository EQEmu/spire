package eqemuserver

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

// pkill takes a regular expression. An empty build tool used to become
// pkill -9 "." (filepath.Base("") is "."), which killed every process of the
// user running Spire: world, zones and Spire itself.
func TestBuildCancelRefusesUnknownBuildTools(t *testing.T) {
	for _, tool := range []string{"", ".", "/", ".*", "nin.a", "make|bash", "/usr/bin/bash", "sh", "ccache"} {
		if targets, err := buildCancelTargets(tool); err == nil {
			t.Errorf("build tool %q: want an error, got targets %v", tool, targets)
		}
	}
}

func TestBuildCancelTargetsKnownBuildTools(t *testing.T) {
	for tool, want := range map[string]string{
		"/usr/bin/make":      "make ccache",
		"/usr/bin/gmake":     "gmake ccache",
		"/usr/bin/ninja":     "ninja ccache",
		"ninja":              "ninja ccache",
		" /usr/bin/ninja \r": "ninja ccache",
	} {
		targets, err := buildCancelTargets(tool)
		if err != nil {
			t.Errorf("build tool %q: %v", tool, err)
			continue
		}
		if got := strings.Join(targets, " "); got != want {
			t.Errorf("build tool %q: targets %q, want %q", tool, got, want)
		}
	}
}

// The name is matched exactly (-x), never as a pattern, and only among this
// user's processes.
func TestBuildCancelKillArgsMatchExactly(t *testing.T) {
	got := strings.Join(buildCancelKillArgs("ninja"), " ")
	want := "-9 -x -u " + strconv.Itoa(os.Getuid()) + " ninja"
	if got != want {
		t.Errorf("pkill args %q, want %q", got, want)
	}
}

// build and build-clean run the build tool the request names; only a known
// build program may run, never an arbitrary command.
func TestBuildToolPathOnlyKnownBuildTools(t *testing.T) {
	for _, tool := range []string{"", ".", "/bin/sh", "bash", "/usr/bin/ninja -t", "../make2"} {
		if path, err := buildToolPath(tool); err == nil {
			t.Errorf("build tool %q: want an error, got %q", tool, path)
		}
	}
	for tool, want := range map[string]string{"/usr/bin/ninja": "/usr/bin/ninja", " make\n": "make", "/usr/local/bin/gmake": "/usr/local/bin/gmake"} {
		path, err := buildToolPath(tool)
		if err != nil || path != want {
			t.Errorf("build tool %q: got %q, %v; want %q", tool, path, err, want)
		}
	}
}
