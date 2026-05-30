package seed

import (
	"embed"
	"errors"
	"fmt"
	"strings"

	"go.yaml.in/yaml/v4"
)

//go:embed manifest/csv_workbooks.yaml
var manifestFS embed.FS

const defaultManifestFile = "manifest/csv_workbooks.yaml"

// ErrInvalidManifest is returned when the CSV workbook manifest violates one of
// the invariants the seeder relies on (e.g. duplicate or empty seedKeys).
var ErrInvalidManifest = errors.New("invalid csv workbook manifest")

// CSVWorkbookManifest maps GCS CSV objects to public workbooks. The workbook
// metadata (title, description, seedKey, language) lives here in the repo; the
// question data lives in the CSV object in GCS and may be appended to over time.
type CSVWorkbookManifest struct {
	Workbooks []CSVWorkbookEntry `yaml:"csvWorkbooks"`
}

// CSVWorkbookEntry describes a single CSV-sourced public workbook.
//
// Format determines how the CSV rows are parsed and which question type they
// produce. SourceLang/TargetLang are the languages of the prompt and the blank
// sentence respectively (they must differ, as word_fill requires).
type CSVWorkbookEntry struct {
	SeedKey     string `yaml:"seedKey"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Language    string `yaml:"language"`
	Format      string `yaml:"format"`
	SourceLang  string `yaml:"sourceLang"`
	TargetLang  string `yaml:"targetLang"`
	GCSObject   string `yaml:"gcsObject"`
}

// DefaultCSVManifest returns the embedded CSV workbook manifest. An empty
// manifest (no entries) is valid and means "no CSV workbooks to seed".
func DefaultCSVManifest() (CSVWorkbookManifest, error) {
	data, err := manifestFS.ReadFile(defaultManifestFile)
	if err != nil {
		return CSVWorkbookManifest{}, fmt.Errorf("read manifest file %s: %w", defaultManifestFile, err)
	}

	var manifest CSVWorkbookManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return CSVWorkbookManifest{}, fmt.Errorf("unmarshal manifest file %s: %w", defaultManifestFile, err)
	}

	if err := manifest.validate(); err != nil {
		return CSVWorkbookManifest{}, fmt.Errorf("validate manifest file %s: %w", defaultManifestFile, err)
	}

	return manifest, nil
}

// validate enforces that seedKeys are present and unique, that the required
// metadata fields are set, and that source and target languages differ.
func (m CSVWorkbookManifest) validate() error {
	seen := make(map[string]bool, len(m.Workbooks))
	for i := range m.Workbooks {
		entry := m.Workbooks[i]
		if entry.SeedKey == "" {
			return fmt.Errorf("csvWorkbook[%d] %q: seedKey must not be empty: %w", i, entry.Title, ErrInvalidManifest)
		}
		if seen[entry.SeedKey] {
			return fmt.Errorf("csvWorkbook[%d]: duplicate seedKey %q: %w", i, entry.SeedKey, ErrInvalidManifest)
		}
		seen[entry.SeedKey] = true

		if entry.Title == "" {
			return fmt.Errorf("csvWorkbook %q: title must not be empty: %w", entry.SeedKey, ErrInvalidManifest)
		}
		if entry.Format == "" {
			return fmt.Errorf("csvWorkbook %q: format must not be empty: %w", entry.SeedKey, ErrInvalidManifest)
		}
		if entry.GCSObject == "" {
			return fmt.Errorf("csvWorkbook %q: gcsObject must not be empty: %w", entry.SeedKey, ErrInvalidManifest)
		}
		if strings.TrimSpace(entry.SourceLang) == "" || strings.TrimSpace(entry.TargetLang) == "" {
			return fmt.Errorf("csvWorkbook %q: sourceLang and targetLang are required: %w", entry.SeedKey, ErrInvalidManifest)
		}
		if entry.SourceLang == entry.TargetLang {
			return fmt.Errorf("csvWorkbook %q: sourceLang and targetLang must differ: %w", entry.SeedKey, ErrInvalidManifest)
		}
	}
	return nil
}
