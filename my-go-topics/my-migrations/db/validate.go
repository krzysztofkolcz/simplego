package db

import (
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Migration struct {
	Version int
	Name    string
	Up      bool
	Down    bool
	FileUp  string
	FileDown string
}


func ValidateMigrations(fsys fs.FS, path string) error {
	var migrationRegex = regexp.MustCompile(`^(\d+)_(.+)\.(up|down)\.sql$`)
	files, err := fs.ReadDir(fsys, path)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("no migration files in %s", path)
	}

	migrations := map[int]*Migration{}

	for _, f := range files {
		name := f.Name()

		// 1. whitespace check
		if strings.Contains(name, " ") {
			return fmt.Errorf("invalid filename (contains space): %s", name)
		}

		// 2. regex match
		matches := migrationRegex.FindStringSubmatch(name)
		if matches == nil {
			return fmt.Errorf("invalid migration filename: %s", name)
		}

		versionStr := matches[1]
		migName := matches[2]
		direction := matches[3]

		version, err := strconv.Atoi(versionStr)
		if err != nil {
			return fmt.Errorf("invalid version in %s", name)
		}

		m, exists := migrations[version]
		if !exists {
			m = &Migration{
				Version: version,
				Name:    migName,
			}
			migrations[version] = m
		}

		// 3. duplicate detection
		if direction == "up" {
			if m.Up {
				return fmt.Errorf("duplicate up migration for version %d", version)
			}
			m.Up = true
			m.FileUp = name
		} else {
			if m.Down {
				return fmt.Errorf("duplicate down migration for version %d", version)
			}
			m.Down = true
			m.FileDown = name
		}

		// 4. empty file check
		content, err := fs.ReadFile(fsys, path+"/"+name)
		if err != nil {
			return err
		}
		if len(strings.TrimSpace(string(content))) == 0 {
			return fmt.Errorf("empty migration file: %s", name)
		}
	}

	// 5. validate pairs
	var versions []int
	for v, m := range migrations {
		if !m.Up {
			return fmt.Errorf("missing up migration for version %d", v)
		}
		if !m.Down {
			return fmt.Errorf("missing down migration for version %d", v)
		}
		versions = append(versions, v)
	}

	sort.Ints(versions)

	// 6. gap detection (opcjonalne — możesz wyłączyć)
	for i := 1; i < len(versions); i++ {
		if versions[i] != versions[i-1]+1 {
			return fmt.Errorf("gap in migrations: %d -> %d", versions[i-1], versions[i])
		}
	}

	return nil
}