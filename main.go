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
func (s LookupPlacesList) String() string {
	var placesStr []string
	for _, p := range s {
		placesStr = append(placesStr, string(p))
	}
	return strings.Join(placesStr, ",")
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
func (s ConfigFileSelector) String() string {
	return fmt.Sprintf("ConfigFileSelector{%v, [%v]}", s.filename, s.lookupPlacesList.String())
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

/* Get first existing configuration file path from added lookup places */
func (s *ConfigFileSelector) SelectFirstKnownPlace() (*string, error) {
	// get first existing configuration file path using possible configuration file path list
	knownPathList, err := s.LookupFilePathList()
	if err != nil {
		return nil, err
	}
	for _, checkPath := range *knownPathList {
		if fileExists, err := s.IsFileExists(checkPath); err == nil && fileExists {
			return &checkPath, nil
		} else {
			continue
		}
	}
	return nil, errors.New(fmt.Sprintf("File not found: %s, path are: %s", s.filename, s.lookupPlacesList))
}

// Find configuration file in requested absolute or relative path.
// Return error if no such file found.
//
// Possible configPath value cases:
//
// - empty string: search filename in defined lookup places only, return first existed or error
// - absolute path: return it if exists or error
// - relative filepath or just filename: search requested in defined lookup places, return first existed or error
func (s *ConfigFileSelector) SelectPath(configPath string) (*string, error) {
	// empty string
	if configPath == "" {
		// config path not set, looking for default
		return s.SelectFirstKnownPlace()
	}
	// absolute path
	if absConfigPath, err := filepath.Abs(configPath); err == nil && absConfigPath == configPath {
		// take absolute path error, cant deal with that
		if exists, err := s.IsFileExists(absConfigPath); err == nil && exists {
			// absolute path ok, return it
			return &absConfigPath, err // err==nil
		}
	}
	// relative filepath, search in defined lookup places but with specified name
	s.filename = configPath
	return s.SelectFirstKnownPlace()
}

/* Check if file specified by full path is exists*/
func (s *ConfigFileSelector) IsFileExists(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	if fileInfo.IsDir() {
		return false, errors.New(fmt.Sprintf("Path %s is dir", path))
	}
	return true, err
}
