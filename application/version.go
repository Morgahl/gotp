package application

import "fmt"

type Version struct {
	major, minor, patch uint
	suffix              string
}

func Ver(major, minor, patch uint, suffix string) Version {
	return Version{major, minor, patch, suffix}
}

func ParseVersion(s string) (Version, error) {
	var v Version
	_, err := fmt.Sscanf(s, "%d.%d.%d-%s", &v.major, &v.minor, &v.patch, &v.suffix)
	if err != nil {
		_, err = fmt.Sscanf(s, "%d.%d.%d", &v.major, &v.minor, &v.patch)
	}
	return v, err
}

func (v Version) Major() uint {
	return v.major
}

func (v Version) Minor() uint {
	return v.minor
}

func (v Version) Patch() uint {
	return v.patch
}

func (v Version) Suffix() string {
	return v.suffix
}

func (v *Version) StepMajor() {
	v.major += 1
	v.minor = 0
	v.patch = 0
	v.suffix = ""
}

func (v *Version) StepMinor() {
	v.minor += 1
	v.patch = 0
	v.suffix = ""
}

func (v *Version) StepPatch() {
	v.patch += 1
	v.suffix = ""
}

func (v *Version) SetSuffix(suffix string) {
	v.suffix = suffix
}

// Less returns true if this version is less than the compared version. It sequentially compares the
// major, minor, patch, and suffix components of the versions; this means that all else equal the
// suffix is compared lexicographically.
func (v Version) Less(other Version) bool {
	switch {
	case v.major != other.major:
		return v.major < other.major
	case v.minor != other.minor:
		return v.minor < other.minor
	case v.patch != other.patch:
		return v.patch < other.patch
	default:
		return v.suffix < other.suffix
	}
}

// Equal compares this version with another version and returns true if this version is equal
func (v Version) Equal(other Version) bool {
	return v.major == other.major && v.minor == other.minor && v.patch == other.patch && v.suffix == other.suffix
}

func (v Version) String() string {
	switch v.suffix {
	case "":
		return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
	default:
		return fmt.Sprintf("%d.%d.%d-%s", v.major, v.minor, v.patch, v.suffix)
	}
}
