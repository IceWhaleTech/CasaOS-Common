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

func Compare(v1, v2 string) (int, error) {
	a1, a2, a3, a4, aa, err := ParseVersion(v1)
	if err != nil {
		return 0, err
	}

	b1, b2, b3, b4, ba, err := ParseVersion(v2)
	if err != nil {
		return 0, err
	}

	if a1 > b1 {
		return 1, nil
	}

	if a1 < b1 {
		return -1, nil
	}

	if a2 > b2 {
		return 1, nil
	}

	if a2 < b2 {
		return -1, nil
	}

	if a3 > b3 {
		return 1, nil
	}

	if a3 < b3 {
		return -1, nil
	}

	if a4 > b4 {
		return 1, nil
	}

	if a4 < b4 {
		return -1, nil
	}

	if aa > ba {
		return 1, nil
	}

	if aa < ba {
		return -1, nil
	}

	return 0, nil
}
