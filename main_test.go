package config_selector

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNewConfigFileSelectorWithNoLookups(t *testing.T) {
	s := NewConfigFileSelector("exampleFilename")
	rType := reflect.TypeOf(s)
	eType := reflect.TypeOf(&ConfigFileSelector{})
	if rType != eType {
		t.Fatalf("Expected NewConfigFileSelector()->%v, not %v", eType.Name(), rType.Name())
	}
	if len(s.lookupPlacesFlags) != 0 {
		t.Fatalf("Unexpected lookup count with no lookups in constructor in %v", s)
	}
}

func TestNewConfigFileSelectorWithLookups(t *testing.T) {
	s := NewConfigFileSelector("exampleFilename", UserConfig, HomeDir)
	rType := reflect.TypeOf(s)
	eType := reflect.TypeOf(&ConfigFileSelector{})
	if rType != eType {
		t.Fatalf("Expected NewConfigFileSelector()->%v, not %v", eType.Name(), rType.Name())
	}
}

func TestConfigFileSelector_AddLookupPlace(t *testing.T) {
	s := NewConfigFileSelector("exampleFilename")
	if len(s.lookupPlacesFlags) != 0 {
		t.Fatalf("Unexpected lookup count with no lookups in constructor in %v", s)
	}
	s.AddLookupPlace(UserConfig)
	if len(s.lookupPlacesFlags) != 1 {
		t.Fatalf("Unexpected lookup count after AddLookupPlace in %v", s)
	}
	s.AddLookupPlace(HomeDir)
	if len(s.lookupPlacesFlags) != 2 {
		t.Fatalf("Unexpected lookup count after AddLookupPlace in %v", s)
	}
}

func TestConfigFileSelector_AddLookupPlace_Uniq(t *testing.T) {
	s := NewConfigFileSelector("exampleFilename")
	if len(s.lookupPlacesFlags) != 0 {
		t.Fatalf("Unexpected lookup count with no lookups in constructor in %v", s)
	}
	s.AddLookupPlace(UserConfig)
	if len(s.lookupPlacesFlags) != 1 {
		t.Fatalf("Unexpected lookup count after AddLookupPlace in %v", s)
	}
	s.AddLookupPlace(UserConfig)
	if len(s.lookupPlacesFlags) != 1 {
		t.Fatalf("Unexpected lookup count after AddLookupPlace in %v", s)
	}
}

func TestConfigFileSelector_LookupFolderList(t *testing.T) {
	type testData struct {
		name    string
		lookup  LookupPlace
		want    *string
		wantErr error
	}
	cwd, _ := os.Getwd()
	home, _ := os.UserHomeDir()

	localTest := ".tmp"
	localTestPath, _ := filepath.Abs(filepath.Join(cwd, localTest))
	fakeFilename := "test.config"
	tests := []testData{
		{"cwd", "./", &cwd, nil},
		{"home", HomeDir, &home, nil},
		{".tmp", LookupPlace(localTest), &localTestPath, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewConfigFileSelector(fakeFilename, tt.lookup)
			if knownDirList, err := s.LookupFolderList(); err != tt.wantErr {
				t.Fatalf("LookupFolderList() unexpected error = %v", err)
			} else if knownDirList == nil {
				t.Fatalf("LookupFolderList() unexpected result %v with error %v", knownDirList, err)
			} else if len(*knownDirList) == 0 {
				t.Fatalf("LookupFolderList() unexpected empty result while expects %#v with error %v", tt.want, err)
			} else {
				var wantStr string
				for _, knownDir := range *knownDirList {

					if tt.want != nil {
						wantStr = *tt.want
					} else {
						wantStr = "<nil>"
					}
					if knownDir != wantStr {
						t.Fatalf("LookupFolderList() got %v instead %v as %v lookup result", knownDir, wantStr, tt.lookup)
					}
					break
				}
			}
		})
	}
}
