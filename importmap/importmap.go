package importmap

import (
	"encoding/json"
	"errors"
	"os"
)

type ImportMap struct {
	Imports map[string]string            `json:"imports"`
	Scopes  map[string]map[string]string `json:"scopes"`
}

func loadImportMapFromPath(path string) (*ImportMap, error) {
	importMap := &ImportMap{}
	file, err := os.Open(path)
	if err != nil {
		return importMap, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(importMap)
	if err != nil {
		return importMap, err
	}
	return importMap, nil
}

func LoadImportMap() (*ImportMap, error) {
	possiblePaths := []string{
		"import-map.json",
		"importMap.json",
		"importmap.json",
		"import_map.json",
	}
	importMap := &ImportMap{}
	for _, path := range possiblePaths {
		importMap, err := loadImportMapFromPath(path)
		if err != nil {
			continue
		} else {
			return importMap, err
		}
	}
	return importMap, errors.New("no import map found")
}

func (m *ImportMap) Get(importPath string) (string, error) {
	if m.Imports == nil {
		return "", errors.New("no imports in import map")
	}
	if m.Imports[importPath] == "" {
		return "", errors.New("import not found in import map")
	}
	return m.Imports[importPath], nil
}

func (m *ImportMap) Has(importPath string) bool {
	if m.Imports == nil {
		return false
	}
	if m.Imports[importPath] == "" {
		return false
	}
	return true
}

func (m *ImportMap) Add(importPath string, importURL string) {
	if m.Imports == nil {
		m.Imports = make(map[string]string)
	}
	m.Imports[importPath] = importURL
}

func (m *ImportMap) Remove(importPath string) {
	if m.Imports == nil {
		return
	}
	delete(m.Imports, importPath)
}

func (m *ImportMap) String() string {
	bytes, _ := json.Marshal(m)
	return string(bytes)
}
