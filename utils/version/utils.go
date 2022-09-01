package version

import (
	"strconv"
	"strings"
)

// retuns major, minor, patch, build and annotation.
//
// e.g. "1.2.3.4-alpha1" returns 1, 2, 3, 4, "alpha1"
func ParseVersion(version string) (int, int, int, int, string, error) {
	ver := strings.TrimSpace(version)
	ver = strings.TrimLeft(ver, "v")

	// get annotation, e.g. "alpha1" in "v0.3.0-alpha1"
	split := strings.SplitN(ver, "-", 2)

	annotation := ""
	if len(split) > 1 {
		annotation = split[1]
	}

	// get version numbers
	va := make([]int, 4)

	for i, s := range strings.Split(split[0], ".") {
		if i >= len(va) {
			break
		}

		v, err := strconv.Atoi(s)
		if err != nil {
			return -1, -1, -1, -1, "", err
		}

		va[i] = v
	}

	return va[0], va[1], va[2], va[3], annotation, nil
}
