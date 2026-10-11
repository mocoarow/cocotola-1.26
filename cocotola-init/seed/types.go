package seed

// PublicWorkbookSeed describes one workbook to be created in the public space.
// SeedKey is the idempotency key and must stay stable across releases.
type PublicWorkbookSeed struct {
	SeedKey     string         `yaml:"seedKey"`
	Title       string         `yaml:"title"`
	Description string         `yaml:"description"`
	Language    string         `yaml:"language"`
	Questions   []QuestionSeed `yaml:"questions"`
}

// QuestionSeed describes a single question seed within a workbook seed.
// SeedKey must be unique within the parent workbook.
type QuestionSeed struct {
	SeedKey      string   `yaml:"seedKey"`
	QuestionType string   `yaml:"questionType"`
	Content      string   `yaml:"content"`
	Tags         []string `yaml:"tags"`
	OrderIndex   int32    `yaml:"orderIndex"`
}
