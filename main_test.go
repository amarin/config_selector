package config_selector

import (
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
