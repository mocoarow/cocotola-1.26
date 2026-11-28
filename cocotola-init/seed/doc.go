// Package seed provides idempotent seeding of public workbooks (and their
// questions) into cocotola-question through WorkbookAPIClient.
//
// The package is intentionally split from the auth-side bootstrap logic so
// that cocotola-init does not depend on cocotola-question's internal data
// layout. All persistence flows through the published internal endpoints,
// preserving the visibility/RBAC invariants enforced inside the workbook
// usecases.
//
// Idempotency is keyed off the workbook's seedKey field (and a seed tag on each
// question), not on the human-friendly title.
// This lets operators rename a workbook in the seed file without producing
// duplicates on the next run.
package seed
