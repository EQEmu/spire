package eqemuserver

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// knownBuildTools are the build programs (CMAKE_MAKE_PROGRAM) that build,
// build-clean and build-cancel may run or stop.
var knownBuildTools = map[string]bool{"make": true, "gmake": true, "ninja": true}

// buildToolPath returns the build tool to run, if it is a known one
// (make, gmake or ninja, by name or full path).
func buildToolPath(buildTool string) (string, error) {
	buildTool = strings.TrimSpace(buildTool)
	if buildTool == "" || !knownBuildTools[filepath.Base(buildTool)] {
		return "", fmt.Errorf("unknown build tool %q (expected make, gmake or ninja)", buildTool)
	}
	return buildTool, nil
}

// buildCancelTargets returns the process names a build cancel stops: the
// build tool and ccache.
//
// pkill matches a regular expression, so the name must never reach it
// unchecked: an empty build tool used to become `pkill -9 .` (filepath.Base("")
// is "."), which killed every process of the user running Spire.
func buildCancelTargets(buildTool string) ([]string, error) {
	path, err := buildToolPath(buildTool)
	if err != nil {
		return nil, err
	}
	return []string{filepath.Base(path), "ccache"}, nil
}

// buildCancelKillArgs are pkill's arguments for one process name: an exact
// name match (-x), only among this user's processes (-u).
func buildCancelKillArgs(name string) []string {
	return []string{"-9", "-x", "-u", strconv.Itoa(os.Getuid()), name}
}
