package config_selector

import (
	"fmt"
	"io/ioutil"
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
	if len(s.lookupPlacesList) != 0 {
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
	if len(s.lookupPlacesList) != 0 {
		t.Fatalf("Unexpected lookup count with no lookups in constructor in %v", s)
	}
	s.AddLookupPlace(UserConfig)
	if len(s.lookupPlacesList) != 1 {
		t.Fatalf("Unexpected lookup count after AddLookupPlace in %v", s)
	}
	s.AddLookupPlace(HomeDir)
	if len(s.lookupPlacesList) != 2 {
		t.Fatalf("Unexpected lookup count after AddLookupPlace in %v", s)
	}
}

func TestConfigFileSelector_AddLookupPlace_Uniq(t *testing.T) {
	s := NewConfigFileSelector("exampleFilename")
	if len(s.lookupPlacesList) != 0 {
		t.Fatalf("Unexpected lookup count with no lookups in constructor in %v", s)
	}
	s.AddLookupPlace(UserConfig)
	if len(s.lookupPlacesList) != 1 {
		t.Fatalf("Unexpected lookup count after AddLookupPlace in %v", s)
	}
	s.AddLookupPlace(UserConfig)
	if len(s.lookupPlacesList) != 1 {
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

func TestConfigFileSelector_UseEtc(t *testing.T) {
	fakeFilename := "test.config"
	s := NewConfigFileSelector(fakeFilename)
	s.UseEtc()
	if knownDirList, err := s.LookupFolderList(); err != nil {
		t.Fatalf("LookupFolderList() unexpected error = %v", err)
	} else if knownDirList == nil {
		t.Fatalf("LookupFolderList() unexpected result %v with error %v", knownDirList, err)
	} else if len(*knownDirList) == 0 {
		t.Fatalf("LookupFolderList() unexpected empty result while expects /etc with error %v", err)
	} else {
		for _, knownDir := range *knownDirList {
			if knownDir != "/etc" {
				t.Fatalf("LookupFolderList() got %v instead /etc while use UseEtc", knownDir)
			}
			break
		}
	}
}

func TestConfigFileSelector_UseEtcProgramFolder(t *testing.T) {
	fakeFilename := "test.config"
	fakeProgramName := "anyName"
	expectedPath := filepath.Join("/etc", fakeProgramName)
	s := NewConfigFileSelector(fakeFilename)
	s.UseEtcProgramFolder(fakeProgramName)
	if knownDirList, err := s.LookupFolderList(); err != nil {
		t.Fatalf("LookupFolderList() unexpected error = %v", err)
	} else if knownDirList == nil {
		t.Fatalf("LookupFolderList() unexpected result %v with error %v", knownDirList, err)
	} else if len(*knownDirList) == 0 {
		t.Fatalf("LookupFolderList() unexpected empty result while expects /etc with error %v", err)
	} else {
		for _, knownDir := range *knownDirList {
			if knownDir != expectedPath {
				t.Fatalf("LookupFolderList() got %v instead %v while use UseEtcProgramFolder", knownDir, expectedPath)
			}
			break
		}
	}
}

type pathType int

const (
	isAbsPath pathType = iota
	isRelPath
	isJustName
	isEmpty
)

type selectPathTestData struct {
	name      string
	path      pathType
	cwdExists bool
	cwdLookup bool
	relExists bool
	relLookup bool
	want      *string
	wantErr   bool
}

func (d selectPathTestData) String() string {
	return fmt.Sprintf("cwd %v rel %v", d.cwdExists, d.relExists)
}

func TestConfigFileSelector_SelectFirstKnownPlace(t *testing.T) {
	absPath, err := ioutil.TempFile("", "test*.conf")
	if err != nil {
		t.Fatalf("Cant create temp file for test: %v", err)
	}
	absFilePath := absPath.Name()
	absFileName := filepath.Base(absFilePath)
	absLookupPlace := LookupPlace(filepath.Dir(absFilePath))

	absRelativePlace := LookupPlace(filepath.Dir(filepath.Dir(absFilePath)))
	absRelativeName := filepath.Join(filepath.Base(filepath.Dir(absFilePath)), absFileName)

	tests := []struct {
		name        string
		filename    string
		lookupPlace *LookupPlace
		exists      bool
		want        *string
		wantErr     bool
	}{
		{"no_lookup_places", absFileName, nil, false, nil, true},
		{"missed", absFileName, &absLookupPlace, false, nil, true},
		{"exists", absFileName, &absLookupPlace, true, &absFilePath, false},
		{"relative/exists", absRelativeName, &absRelativePlace, true, &absFilePath, false},
		{"relative/missed", absRelativeName, &absRelativePlace, false, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searchFileName := tt.filename
			if !tt.exists {
				searchFileName += ".really.not.exists"
			}
			s := &ConfigFileSelector{
				filename: searchFileName,
			}
			if tt.lookupPlace != nil {
				s.AddLookupPlace(*tt.lookupPlace)
			}
			got, err := s.SelectFirstKnownPlace()
			gotStr := "<nil>"
			wantStr := "<nil>"
			if got != nil {
				gotStr = *got
			}
			if tt.want != nil {
				wantStr = *tt.want
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectFirstKnownPlace()=(%v, %v), wantErr %v at %v", gotStr, err, tt.wantErr, s)
				return
			}
			if gotStr != wantStr {
				t.Fatalf("SelectFirstKnownPlace()=(%v, nil) != %v for exists %v with %v", gotStr, wantStr, searchFileName, s)
			}
		})
	}
}

func TestConfigFileSelector_SelectPath(t *testing.T) {
	baseFolder, err := ioutil.TempDir("", "sp")
	if err != nil {
		t.Fatalf("Cant create temp path test: %v", err)
	}
	absFolder, err := ioutil.TempDir(baseFolder, "existed")
	if err != nil {
		t.Fatalf("Cant create temp path for temp config: %v", err)
	}
	absPath, err := ioutil.TempFile(absFolder, "test*.conf")
	if err != nil {
		t.Fatalf("Cant create temp file for test: %v", err)
	}
	absFilePath := absPath.Name()
	absFileName := filepath.Base(absFilePath)
	cwdFolder, err := os.Getwd()
	if err != nil {
		t.Fatalf("Cant detect current work dir: %v", err)
	}
	cwdFilePath := filepath.Join(cwdFolder, absFileName)
	relFolder, err := ioutil.TempDir(baseFolder, "rel")
	if err != nil {
		t.Fatalf("Cant create temp path for temp config: %v", err)
	}
	relFolderName := filepath.Base(relFolder)
	relParent := filepath.Dir(relFolder)
	relFilePath := filepath.Join(relFolder, absFileName)

	tests := []selectPathTestData{
		{
			"empty", isEmpty,
			false, true,
			false, true,
			nil, true,
		},
		{
			"absolutePath", isAbsPath,
			true, true,
			true, true,
			&absFilePath, false,
		},
		{
			"./existed", isJustName,
			true, true,
			false, true,
			&cwdFilePath, false,
		},
		{
			"rel/missed", isRelPath,
			true, true,
			false, false,
			nil, true,
		},
		{
			"rel/existed", isRelPath,
			true, true,
			true, true,
			&relFilePath, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewConfigFileSelector(absFileName)
			// remove old files
			if exists, err := s.IsFileExists(cwdFilePath); err != nil {
				t.Fatalf("Cant check if %v exists in cwd %s", absFileName, cwdFolder)
			} else if exists {
				if err := os.Remove(cwdFilePath); err != nil {
					t.Fatalf("Cant remove cwd link %s", cwdFilePath)
				}
			}
			if exists, err := s.IsFileExists(relFilePath); err != nil {
				t.Fatalf("Cant check if %v exists in %s", absFileName, relFolder)
			} else if exists {
				if err := os.Remove(relFilePath); err != nil {
					t.Fatalf("Cant remove cwd link %s", relFilePath)
				}
			}
			// add lookup places if required
			if tt.cwdLookup {
				s.AddLookupPlace(LookupPlace(cwdFolder))
			}
			if tt.relLookup {
				s.AddLookupPlace(LookupPlace(relParent))
			}
			// make cwd & rel files if required
			if tt.cwdExists {
				if err := os.Symlink(absFilePath, cwdFilePath); err != nil {
					t.Fatalf("Cant symlink test config to cwd %v", cwdFolder)
				}
			}
			if tt.relExists {
				if err := os.Symlink(absFilePath, relFilePath); err != nil {
					t.Fatalf("Cant symlink test config to %v", relFolder)
				}
			}
			var got *string
			var err error
			var selectPath string

			switch tt.path {
			case isEmpty:
				selectPath = ""
			case isAbsPath:
				selectPath = absFilePath
			case isJustName:
				selectPath = absFileName
			case isRelPath:
				selectPath = filepath.Join(relFolderName, absFileName)
			}
			got, err = s.SelectPath(selectPath)
			gotStr := "<nil>"
			wantStr := "<nil>"
			if got != nil {
				gotStr = *got
			}
			if tt.want != nil {
				wantStr = *tt.want
			}
			if tt.wantErr && err == nil {
				t.Errorf("SelectPath(%#v) = (%v, %v) for %v with %v", selectPath, gotStr, err, tt, s)
			}

			if gotStr != wantStr {
				t.Errorf("SelectPath(%v)=(%v, %v) != %v for %v with %v", selectPath, gotStr, err, wantStr, tt, s)
			}
			if exists, err := s.IsFileExists(cwdFilePath); err != nil {
				t.Fatalf("Cant check if %v exists in cwd %s", absFileName, cwdFolder)
			} else if exists {
				if err := os.Remove(cwdFilePath); err != nil {
					t.Fatalf("Cant remove cwd link %s", cwdFilePath)
				}
			}
			if exists, err := s.IsFileExists(relFilePath); err != nil {
				t.Fatalf("Cant check if %v exists in %s", absFileName, relFolder)
			} else if exists {
				if err := os.Remove(relFilePath); err != nil {
					t.Fatalf("Cant remove cwd link %s", relFilePath)
				}
			}
		})
	}
}

func TestLookupPlacesList_String(t *testing.T) {
	tests := []struct {
		name string
		s    LookupPlacesList
		want string
	}{
		{"empty", LookupPlacesList{}, ""},
		{"home", LookupPlacesList{HomeDir}, fmt.Sprintf("%s", HomeDir)},
		{"etc", LookupPlacesList{Etc}, fmt.Sprintf("%s", Etc)},
		{"home,etc", LookupPlacesList{HomeDir, Etc}, fmt.Sprintf("%s,%s", HomeDir, Etc)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigFileSelector_GetLookupPlaces_String(t *testing.T) {
	tests := []struct {
		name string
		s    LookupPlacesList
		want string
	}{
		{"empty", LookupPlacesList{}, ""},
		{"home", LookupPlacesList{HomeDir}, fmt.Sprintf("%s", HomeDir)},
		{"etc", LookupPlacesList{Etc}, fmt.Sprintf("%s", Etc)},
		{"home,etc", LookupPlacesList{HomeDir, Etc}, fmt.Sprintf("%s,%s", HomeDir, Etc)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := NewConfigFileSelector("any")
			for _, lp := range tt.s {
				cs.AddLookupPlace(lp)
			}
			if got := fmt.Sprintf("%s", cs.GetLookupPlaces()); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
