package seed

// ValidateSeeds exposes validateSeeds to the external seed_test package.
var ValidateSeeds = validateSeeds

// ValidateCSVManifest exposes CSVWorkbookManifest.validate to the external seed_test package.
func ValidateCSVManifest(m CSVWorkbookManifest) error {
	return m.validate()
}
