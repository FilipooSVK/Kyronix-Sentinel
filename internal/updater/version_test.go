package updater

import "testing"

func TestNormalizeVersion(
	t *testing.T,
) {

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "0.1.0",
			expected: "v0.1.0",
		},
		{
			input:    "v0.1.0",
			expected: "v0.1.0",
		},
		{
			input:    " 0.2.0 ",
			expected: "v0.2.0",
		},
	}

	for _, test := range tests {

		got := NormalizeVersion(
			test.input,
		)

		if got != test.expected {

			t.Fatalf(
				"expected %s, got %s",
				test.expected,
				got,
			)
		}
	}
}

func TestIsValidVersion(
	t *testing.T,
) {

	if !IsValidVersion(
		"0.1.0",
	) {

		t.Fatal(
			"expected 0.1.0 to be valid",
		)
	}

	if IsValidVersion(
		"not-a-version",
	) {

		t.Fatal(
			"expected invalid version",
		)
	}
}

func TestIsNewerVersion(
	t *testing.T,
) {

	if !IsNewerVersion(
		"0.1.0",
		"0.1.1",
	) {

		t.Fatal(
			"expected 0.1.1 to be newer than 0.1.0",
		)
	}
}

func TestIsNewerVersionSameVersion(
	t *testing.T,
) {

	if IsNewerVersion(
		"0.1.0",
		"v0.1.0",
	) {

		t.Fatal(
			"same version must not be reported as newer",
		)
	}
}

func TestIsNewerVersionOlderCandidate(
	t *testing.T,
) {

	if IsNewerVersion(
		"0.2.0",
		"0.1.9",
	) {

		t.Fatal(
			"older candidate must not be reported as newer",
		)
	}
}

func TestIsNewerVersionMajorUpgrade(
	t *testing.T,
) {

	if !IsNewerVersion(
		"0.9.9",
		"1.0.0",
	) {

		t.Fatal(
			"expected 1.0.0 to be newer than 0.9.9",
		)
	}
}

func TestPrereleaseVersionOrdering(
	t *testing.T,
) {

	if !IsNewerVersion(
		"0.2.0-beta.1",
		"0.2.0",
	) {

		t.Fatal(
			"stable release must be newer than prerelease",
		)
	}

	if IsNewerVersion(
		"0.2.0",
		"0.2.0-beta.1",
	) {

		t.Fatal(
			"prerelease must not replace stable release",
		)
	}
}

func TestInvalidCandidateVersion(
	t *testing.T,
) {

	if IsNewerVersion(
		"0.1.0",
		"latest",
	) {

		t.Fatal(
			"invalid candidate must not be accepted",
		)
	}
}
