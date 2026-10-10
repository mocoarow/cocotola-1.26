package main

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-init/config"
	"github.com/mocoarow/cocotola-1.26/cocotola-init/initialize"
	"github.com/mocoarow/cocotola-1.26/cocotola-init/seed"
)

func runInit(ctx context.Context, cfg *config.Config, db *gorm.DB) error {
	policyEnsurer, err := initialize.NewWorkbookPolicyEnsurer(db)
	if err != nil {
		return fmt.Errorf("new workbook policy ensurer: %w", err)
	}
	client, err := newQuestionAPIClient(ctx, cfg.AppEnv, cfg.Question)
	if err != nil {
		return fmt.Errorf("new question api client: %w", err)
	}
	seeds, err := seed.DefaultSeeds()
	if err != nil {
		return fmt.Errorf("load default seeds: %w", err)
	}
	seeder := seed.NewWorkbookSeeder(client, policyEnsurer, seeds)

	slog.InfoContext(ctx, "starting initialization", slog.String("app", appName))
	if err := initialize.Initialize(ctx, db, seeder, cfg.App.OwnerLoginID, cfg.App.OwnerPassword); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	slog.InfoContext(ctx, "initialization completed successfully")
	return nil
}
