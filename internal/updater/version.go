package updater

import (
	"strconv"
	"strings"
)

type semanticVersion struct {
	major int

	minor int

	patch int

	prerelease []string
}

// NormalizeVersion normalizes Sentinel release versions.
//
// Examples:
//
// 0.1.0  -> v0.1.0
// v0.1.0 -> v0.1.0
func NormalizeVersion(
	version string,
) string {

	version = strings.TrimSpace(
		version,
	)

	if version == "" {
		return ""
	}

	if !strings.HasPrefix(
		version,
		"v",
	) {

		version = "v" + version
	}

	return version
}

// IsValidVersion reports whether version is a supported semantic version.
func IsValidVersion(
	version string,
) bool {

	_, ok := parseSemanticVersion(
		version,
	)

	return ok
}

// IsNewerVersion reports whether candidate is newer than current.
func IsNewerVersion(
	current string,
	candidate string,
) bool {

	if !IsValidVersion(current) ||
		!IsValidVersion(candidate) {

		return false
	}

	return CompareVersions(
		candidate,
		current,
	) > 0
}

// CompareVersions compares two semantic versions.
//
// Returns:
//
// -1 when first < second
//
//	0 when first == second
//	1 when first > second
//
// Invalid versions compare as equal. Call IsValidVersion when
// validation is required before comparison.
func CompareVersions(
	first string,
	second string,
) int {

	a, okA := parseSemanticVersion(
		first,
	)

	b, okB := parseSemanticVersion(
		second,
	)

	if !okA || !okB {
		return 0
	}

	if a.major != b.major {

		if a.major < b.major {
			return -1
		}

		return 1
	}

	if a.minor != b.minor {

		if a.minor < b.minor {
			return -1
		}

		return 1
	}

	if a.patch != b.patch {

		if a.patch < b.patch {
			return -1
		}

		return 1
	}

	return comparePrerelease(
		a.prerelease,
		b.prerelease,
	)
}

func parseSemanticVersion(
	version string,
) (semanticVersion, bool) {

	version = NormalizeVersion(
		version,
	)

	if version == "" {
		return semanticVersion{}, false
	}

	version = strings.TrimPrefix(
		version,
		"v",
	)

	// Build metadata does not affect version precedence.
	if index := strings.Index(
		version,
		"+",
	); index >= 0 {

		build := version[index+1:]

		if !validIdentifiers(
			build,
			false,
		) {

			return semanticVersion{}, false
		}

		version = version[:index]
	}

	var prerelease []string

	if index := strings.Index(
		version,
		"-",
	); index >= 0 {

		value := version[index+1:]

		if !validIdentifiers(
			value,
			true,
		) {

			return semanticVersion{}, false
		}

		prerelease = strings.Split(
			value,
			".",
		)

		version = version[:index]
	}

	parts := strings.Split(
		version,
		".",
	)

	if len(parts) != 3 {
		return semanticVersion{}, false
	}

	major, ok := parseCoreNumber(
		parts[0],
	)

	if !ok {
		return semanticVersion{}, false
	}

	minor, ok := parseCoreNumber(
		parts[1],
	)

	if !ok {
		return semanticVersion{}, false
	}

	patch, ok := parseCoreNumber(
		parts[2],
	)

	if !ok {
		return semanticVersion{}, false
	}

	return semanticVersion{
		major: major,

		minor: minor,

		patch: patch,

		prerelease: prerelease,
	}, true
}

func parseCoreNumber(
	value string,
) (int, bool) {

	if value == "" {
		return 0, false
	}

	if len(value) > 1 &&
		value[0] == '0' {

		return 0, false
	}

	for _, character := range value {

		if character < '0' ||
			character > '9' {

			return 0, false
		}
	}

	number, err := strconv.Atoi(
		value,
	)

	if err != nil {
		return 0, false
	}

	return number, true
}

func validIdentifiers(
	value string,
	checkNumericLeadingZero bool,
) bool {

	if value == "" {
		return false
	}

	identifiers := strings.Split(
		value,
		".",
	)

	for _, identifier := range identifiers {

		if identifier == "" {
			return false
		}

		numeric := true

		for _, character := range identifier {

			if (character < '0' ||
				character > '9') &&
				(character < 'A' ||
					character > 'Z') &&
				(character < 'a' ||
					character > 'z') &&
				character != '-' {

				return false
			}

			if character < '0' ||
				character > '9' {

				numeric = false
			}
		}

		if checkNumericLeadingZero &&
			numeric &&
			len(identifier) > 1 &&
			identifier[0] == '0' {

			return false
		}
	}

	return true
}

func comparePrerelease(
	first []string,
	second []string,
) int {

	// A normal release has higher precedence than a prerelease.
	if len(first) == 0 &&
		len(second) == 0 {

		return 0
	}

	if len(first) == 0 {
		return 1
	}

	if len(second) == 0 {
		return -1
	}

	limit := len(first)

	if len(second) < limit {
		limit = len(second)
	}

	for i := 0; i < limit; i++ {

		result := comparePrereleaseIdentifier(
			first[i],
			second[i],
		)

		if result != 0 {
			return result
		}
	}

	if len(first) < len(second) {
		return -1
	}

	if len(first) > len(second) {
		return 1
	}

	return 0
}

func comparePrereleaseIdentifier(
	first string,
	second string,
) int {

	firstNumber, firstNumeric := numericIdentifier(
		first,
	)

	secondNumber, secondNumeric := numericIdentifier(
		second,
	)

	switch {

	case firstNumeric &&
		secondNumeric:

		if firstNumber < secondNumber {
			return -1
		}

		if firstNumber > secondNumber {
			return 1
		}

		return 0

	case firstNumeric:

		return -1

	case secondNumeric:

		return 1

	default:

		if first < second {
			return -1
		}

		if first > second {
			return 1
		}

		return 0
	}
}

func numericIdentifier(
	value string,
) (int, bool) {

	if value == "" {
		return 0, false
	}

	for _, character := range value {

		if character < '0' ||
			character > '9' {

			return 0, false
		}
	}

	number, err := strconv.Atoi(
		value,
	)

	if err != nil {
		return 0, false
	}

	return number, true
}
