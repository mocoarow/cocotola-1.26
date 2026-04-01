package config

// QuestionConfig holds configuration for the cocotola-question service.
type QuestionConfig struct {
	FirestoreProjectID string `yaml:"firestoreProjectId" validate:"required"`
}
