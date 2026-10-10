package main

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-init/config"
	"github.com/mocoarow/cocotola-1.26/cocotola-init/gateway"
	"github.com/mocoarow/cocotola-1.26/cocotola-init/importer"
	"github.com/mocoarow/cocotola-1.26/cocotola-init/initialize"
	"github.com/mocoarow/cocotola-1.26/cocotola-init/seed"
)

func runImport(ctx context.Context, cfg *config.Config, db *gorm.DB) error {
	policyEnsurer, err := initialize.NewWorkbookPolicyEnsurer(db)
	if err != nil {
		return fmt.Errorf("new workbook policy ensurer: %w", err)
	}
	client, err := newQuestionAPIClient(ctx, cfg.AppEnv, cfg.Question)
	if err != nil {
		return fmt.Errorf("new question api client: %w", err)
	}
	manifest, err := seed.DefaultCSVManifest()
	if err != nil {
		return fmt.Errorf("load csv manifest: %w", err)
	}
	reader, err := gateway.NewGCSReader(ctx, cfg.CSVSeed.BucketName)
	if err != nil {
		return fmt.Errorf("new gcs reader: %w", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			slog.WarnContext(ctx, "close gcs reader", slog.Any("error", closeErr))
		}
	}()

	im := importer.NewImporter(
		initialize.CocotolaOrganizationID(),
		initialize.NewPublicSpaceFinder(db),
		seed.NewCSVWorkbookLoader(reader, manifest),
		seed.NewWorkbookSeeder(client, policyEnsurer, nil),
	)
	if err := im.Run(ctx); err != nil {
		return fmt.Errorf("import: %w", err)
	}
	return nil
}
