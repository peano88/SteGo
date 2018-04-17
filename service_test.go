package SteGo

import (
	"fmt"
	"testing"
)

const baseVersions = 2

var s Service

type testCase struct {
	version ServiceVersion
	errorf  bool
	pos     int
}

type testCaseV struct {
	version  int
	expected int
}

func initServiceVersion() []*ServiceVersion {

	var v []*ServiceVersion

	for i := 1; i <= baseVersions; i++ {
		//Create a Version
		sv := ServiceVersion{
			Address: fmt.Sprintf("addr%v", i),
			Version: i,
		}

		v = append(v, &sv)
	}
	return v
}

func Test_Add(t *testing.T) {

	s := Service{
		Versions: initServiceVersion(),
	}

	tc := []testCase{
		//0:Add Version immediately consecutive
		testCase{
			version: ServiceVersion{
				Address: fmt.Sprintf("addr%v", baseVersions+1),
				Version: baseVersions + 1,
			},
			errorf: false,
			pos:    baseVersions,
		},

		//1:Add Version with Gap
		testCase{
			version: ServiceVersion{
				Address: fmt.Sprintf("addr%v", baseVersions+4),
				Version: baseVersions + 4,
			},
			errorf: false,
			pos:    baseVersions + 3,
		},
		//2:Add Version between TestCase 1 and 2
		testCase{
			version: ServiceVersion{
				Address: fmt.Sprintf("addr%v", baseVersions+3),
				Version: baseVersions + 3,
			},
			errorf: false,
			pos:    baseVersions + 2,
		},
		//3:Add Version without address
		testCase{
			version: ServiceVersion{
				Address: "",
				Version: baseVersions + 4,
			},
			errorf: true,
		},
		//4:Add Version with overlapping version
		testCase{
			version: ServiceVersion{
				Address: "testOverlap",
				Version: baseVersions,
			},
			errorf: true,
		},
	}

	for j, test := range tc {

		err := s.AddVersion(test.version)

		switch {
		case err != nil && test.errorf == false:
			t.Errorf("testCase: %v ::: %v", j, err.Error())
		case err == nil && test.errorf == true:
			t.Errorf("testCase: %v ::: Expected an error", j)
		}

		if test.errorf == true {
			continue
		}

		if s.Versions[test.pos].Version != test.version.Version {
			t.Errorf("testCase: %v ::: Expected : %v found %v", j, test.version.Version, s.Versions[test.pos])
		}

	}

}

func Test_Version(t *testing.T) {

	s := Service{
		Versions: initServiceVersion(),
	}

	// add a service with Gap
	sv := ServiceVersion{
		Address: fmt.Sprintf("addr%v", baseVersions+2),
		Version: baseVersions + 2,
	}

	s.Versions = append(append(s.Versions, nil), &sv)

	tc := []testCaseV{
		//0: Version existing
		testCaseV{
			version:  1,
			expected: 1,
		},
		//1: Version not existing: in the gap
		testCaseV{
			version:  baseVersions + 1,
			expected: baseVersions,
		},
		//2: Version not existing: > len(versions)
		testCaseV{
			version:  baseVersions + 3,
			expected: baseVersions + 2,
		},
	}

	for i, test := range tc {
		candidate := s.Version(test.version)

		if candidate.Version != test.expected {
			t.Errorf("Test case:%v ::: Expected %v found %v", i, test.expected, candidate.Version)
		}

	}
}
