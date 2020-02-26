package config_selector

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type LookupPlace string

const (
	HomeDir     LookupPlace = "Home"
	UserConfig  LookupPlace = ".config"
	CurrentPath LookupPlace = "./"
)

type SearchPlaces []LookupPlace

type ConfigFileSelector struct {
	filename          string
	lookupPlacesFlags SearchPlaces
}

func (s *ConfigFileSelector) String() string {
	var placesStr []string
	for _, p := range s.lookupPlacesFlags {
		placesStr = append(placesStr, string(p))
	}
	return fmt.Sprintf("ConfigFileSelector{%v, [%v]}", s.filename, strings.Join(placesStr, ","))
}

/* Make new configuration loader for required filename using search places flags */
func NewConfigFileSelector(fileName string, a ...LookupPlace) *ConfigFileSelector {
	lookupPlaces := SearchPlaces{}
	for _, plc := range a {
		lookupPlaces = append(lookupPlaces, plc)
	}
	return &ConfigFileSelector{fileName, lookupPlaces}
}

/* Add lookup place */
func (s *ConfigFileSelector) AddLookupPlace(place LookupPlace) {
	for _, p := range s.lookupPlacesFlags {
		if p == place {
			return
		}
	}
	s.lookupPlacesFlags = append(s.lookupPlacesFlags, place)
}

/* Get well-known path list for searching config file's

Returns a list of well-known directories in order set by lookup flags
*/
func (s *ConfigFileSelector) LookupFolderList() (*[]string, error) {
	var lookupPlaces []string
	for _, placeKey := range s.lookupPlacesFlags {
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

// Find configuration file in path or well known path list defined by lookup flags
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
