package config_selector

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LookupPlace type is just a string
type LookupPlace string

// Some lookup places predefined.
// Constant lookup places has special meanings which actual path calculated runtime.
// All others will threat as relative or absolute path
const (
	// Home is a platform dependent user home folder. Actual place detected runtime
	HomeDir LookupPlace = "Home"
	// UserConfig dir is a platform dependent user configuration path. Actual place detected runtime
	UserConfig LookupPlace = ".config"
	// Current work dir
	CurrentPath LookupPlace = "./"
	// Posix platforms /etc/
	Etc LookupPlace = "/etc"
)

// Lookup places stored together in simple slice
type LookupPlacesList []LookupPlace

// LookupPlacesList has it's own String method, it useful to output lookup places list, separated by comma's
func (s *LookupPlacesList) String() string {
	var placesStr []string
	for _, p := range *s {
		placesStr = append(placesStr, string(p))
	}
	return strings.Join(placesStr, ", ")
}

// ConfigFileSelector is an a helper object.
// It provides some methods to easy find required configuration file specified by name
// Literal form initialisation:
//
//   configSelector = &ConfigFileSelector{filename, LookupPlacesList{CurrentPath, HomeDir}}
//
// Additionally you can use NewConfigFileSelector constructor:
//
//   configSelector = NewConfigFileSelector(filename, CurrentPath, HomeDir)
//
type ConfigFileSelector struct {
	filename         string
	lookupPlacesList LookupPlacesList
}

// ConfigFileSelector instance implements Stringer interface:
//
// configSelector.String() == "ConfigFileSelector{filename.conf, [./, Home]}
func (s *ConfigFileSelector) String() string {
	return fmt.Sprintf("ConfigFileSelector{%v, [%v]}", s.filename, s.lookupPlacesList)
}

// Get lookup places keys list.
// Useful to check some place is in list or to call configFileSelector.GetLookupPlaces().String()
func (s *ConfigFileSelector) GetLookupPlaces() LookupPlacesList {
	return s.lookupPlacesList
}

/* ConfigFileSelector constructor
Allow to make new instance passing required filename and set of lookup places objects */
func NewConfigFileSelector(fileName string, a ...LookupPlace) *ConfigFileSelector {
	lookupPlaces := LookupPlacesList{}
	for _, plc := range a {
		lookupPlaces = append(lookupPlaces, plc)
	}
	return &ConfigFileSelector{fileName, lookupPlaces}
}

// AddLookupPlace allow to add additional lookup place runtime
func (s *ConfigFileSelector) AddLookupPlace(place LookupPlace) {
	for _, p := range s.lookupPlacesList {
		if p == place {
			return
		}
	}
	s.lookupPlacesList = append(s.lookupPlacesList, place)
}

// UseEtc adds /etc/ path to ConfigFileSelector lookup places list
func (s *ConfigFileSelector) UseEtc() {
	s.lookupPlacesList = append(s.lookupPlacesList, Etc)
}

// UseEtcProgramFolder allow to add /etc/<program name>/ path to ConfigFileSelector lookup places list
func (s *ConfigFileSelector) UseEtcProgramFolder(programName string) {
	s.lookupPlacesList = append(s.lookupPlacesList, LookupPlace(filepath.Join("/etc", programName)))
}

// LookupFolderList just return list of absolute path string if such path exists.
// All runtime calculated path are resolved here
// Returns a list of well-known directories in order set by lookup flags and consequent additions
func (s *ConfigFileSelector) LookupFolderList() (*[]string, error) {
	var lookupPlaces []string
	for _, placeKey := range s.lookupPlacesList {
		switch placeKey {
		case UserConfig:
			if userConfigDir, err := os.UserConfigDir(); err == nil {
				if absPath, err := filepath.Abs(userConfigDir); err == nil {
					lookupPlaces = append(lookupPlaces, absPath)
				}
			}
		case HomeDir:
			if homeDir, err := os.UserHomeDir(); err == nil {
				if absPath, err := filepath.Abs(homeDir); err == nil {
					lookupPlaces = append(lookupPlaces, absPath)
				}
			}
		case CurrentPath:
			if currentDir, err := os.Getwd(); err == nil {
				if absPath, err := filepath.Abs(currentDir); err == nil {
					lookupPlaces = append(lookupPlaces, absPath)
				}
			}
		default:
			if absPath, err := filepath.Abs(string(placeKey)); err != nil {
				return nil, err
			} else {
				lookupPlaces = append(lookupPlaces, absPath)
			}
		}

	}
	return &lookupPlaces, nil
}

/* Get possible configuration file path list using lookup places in order set by lookup flags */
func (s *ConfigFileSelector) LookupFilePathList() (*[]string, error) {
	// get possible configuration file path using well-known config dir list
	ret := make([]string, 0)
	knownDirList, err := s.LookupFolderList()
	if err != nil {
		return nil, err
	}
	for _, knownDir := range *knownDirList {
		configPathElements := []string{knownDir, s.filename}
		checkPath := strings.Join(configPathElements, string(os.PathSeparator))
		ret = append(ret, checkPath)
	}
	return &ret, nil
}

/* Get first existing configuration file path from places set by lookup flags */
func (s *ConfigFileSelector) SelectFirstKnownPlace() (*string, error) {
	// get first existing configuration file path using possible configuration file path list
	knownPathList, err := s.LookupFilePathList()
	if err != nil {
		return nil, err
	}
	for _, checkPath := range *knownPathList {
		if fileExists, err := s.IsFileExists(checkPath); err == nil && fileExists {
			return &checkPath, nil
		} else if err == nil && !fileExists {
			return nil, err
		} else {
			return nil, err
		}
	}
	return nil, errors.New(fmt.Sprintf(
		"No %v found in %v", s.filename, strings.Join(*knownPathList, ", "),
	))
}

// Find configuration file in requested path first or in well known path list defined by lookup flags
// return error if no such file found either in requested path or in well known path list
func (s *ConfigFileSelector) SelectPath(path string) (*string, error) {
	lookupPath := strings.Join([]string{path, s.filename}, string(os.PathSeparator))
	if fileExists, err := s.IsFileExists(lookupPath); err == nil && fileExists == true {
		return &lookupPath, nil
	} else if err != nil {
		// log lookup error
		return s.SelectFirstKnownPlace()
	} else if fileExists != true {
		// log file not found
		return s.SelectFirstKnownPlace()
	}
	return nil, os.ErrNotExist
}

/* Check if file specified by full path is exists*/
func (s *ConfigFileSelector) IsFileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
