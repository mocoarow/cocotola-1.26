package seed

import (
	"embed"
	"fmt"

	"go.yaml.in/yaml/v4"
)

//go:embed seeds/public_workbooks.yaml
var seedsFS embed.FS

const defaultSeedFile = "seeds/public_workbooks.yaml"

type seedFile struct {
	Workbooks []PublicWorkbookSeed `yaml:"workbooks"`
}

// DefaultSeeds returns the embedded default set of public workbook seeds.
// Failure to parse the embedded YAML is surfaced as an error so callers can
// log it through the normal startup path instead of a bare panic.
func DefaultSeeds() ([]PublicWorkbookSeed, error) {
	seeds, err := loadSeedFile(defaultSeedFile)
	if err != nil {
		return nil, fmt.Errorf("default seed file is invalid: %w", err)
	}
	return seeds, nil
}

func loadSeedFile(path string) ([]PublicWorkbookSeed, error) {
	data, err := seedsFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read seed file %s: %w", path, err)
	}

	var f seedFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("unmarshal seed file %s: %w", path, err)
	}

	if err := validateSeeds(f.Workbooks); err != nil {
		return nil, fmt.Errorf("validate seed file %s: %w", path, err)
	}

	return f.Workbooks, nil
}

// validateSeeds enforces the invariants the seeder relies on:
//   - workbook seedKeys are non-empty and unique across the file
//   - question seedKeys are non-empty and unique within a workbook
func validateSeeds(workbooks []PublicWorkbookSeed) error {
	seenWorkbooks := make(map[string]bool, len(workbooks))
	for i := range workbooks {
		wb := workbooks[i]
		if wb.SeedKey == "" {
			return fmt.Errorf("workbook[%d] %q: seedKey must not be empty", i, wb.Title)
		}
		if seenWorkbooks[wb.SeedKey] {
			return fmt.Errorf("workbook[%d]: duplicate seedKey %q", i, wb.SeedKey)
		}
		seenWorkbooks[wb.SeedKey] = true

		seenQuestions := make(map[string]bool, len(wb.Questions))
		for j, q := range wb.Questions {
			if q.SeedKey == "" {
				return fmt.Errorf("workbook %q question[%d]: seedKey must not be empty", wb.SeedKey, j)
			}
			if seenQuestions[q.SeedKey] {
				return fmt.Errorf("workbook %q question[%d]: duplicate seedKey %q", wb.SeedKey, j, q.SeedKey)
			}
			seenQuestions[q.SeedKey] = true
		}
	}
	return nil
}
