package seed

// PublicWorkbookSeed describes one workbook (and its questions) to be created
// in the public space. SeedKey is the durable, machine-readable identifier
// used for idempotency checks; it must remain stable across releases even if
// Title or Description change.
type PublicWorkbookSeed struct {
	SeedKey     string         `yaml:"seedKey"`
	Title       string         `yaml:"title"`
	Description string         `yaml:"description"`
	Questions   []QuestionSeed `yaml:"questions"`
}

// QuestionSeed describes a single question seed within a workbook seed.
// SeedKey is unique within the parent workbook and is encoded into the tag
// `seed:<workbookSeedKey>:<questionSeedKey>` so the seeder can detect
// existing questions on subsequent runs.
type QuestionSeed struct {
	SeedKey      string   `yaml:"seedKey"`
	QuestionType string   `yaml:"questionType"`
	Content      string   `yaml:"content"`
	Tags         []string `yaml:"tags"`
	OrderIndex   int32    `yaml:"orderIndex"`
}
