package main

import (
	"testing"
)

// TestBuildMigrateStatusReport_AllApplied pins the report shape when
// the running binary is at the latest version (the normal case for
// auto-apply): every step lands in HistorySteps, PendingSteps stays
// empty.
func TestBuildMigrateStatusReport_AllApplied(t *testing.T) {
	got := buildMigrateStatusReport("sqlite://x", currentMigrationVersion())
	if got.CurrentVersion != got.LatestVersion {
		t.Errorf("current %d != latest %d", got.CurrentVersion, got.LatestVersion)
	}
	if len(got.PendingSteps) != 0 {
		t.Errorf("expected no pending steps, got %+v", got.PendingSteps)
	}
	if len(got.HistorySteps) != len(chronosMigrations) {
		t.Errorf("history len %d, want %d", len(got.HistorySteps), len(chronosMigrations))
	}
}

// TestBuildMigrateStatusReport_PendingPath proves the report splits
// the ladder correctly when current < latest. Mirrors the situation
// an operator faces right after pulling a new chronos binary
// against an older database.
func TestBuildMigrateStatusReport_PendingPath(t *testing.T) {
	if len(chronosMigrations) < 2 {
		t.Skip("need ≥ 2 declared migrations for this test to be meaningful")
	}
	at := chronosMigrations[0].Version
	got := buildMigrateStatusReport("sqlite://x", at)
	if got.CurrentVersion != at {
		t.Errorf("current = %d, want %d", got.CurrentVersion, at)
	}
	if got.LatestVersion <= at {
		t.Errorf("latest %d not > current %d", got.LatestVersion, at)
	}
	wantHistory := 1
	wantPending := len(chronosMigrations) - wantHistory
	if len(got.HistorySteps) != wantHistory {
		t.Errorf("history len = %d, want %d", len(got.HistorySteps), wantHistory)
	}
	if len(got.PendingSteps) != wantPending {
		t.Errorf("pending len = %d, want %d", len(got.PendingSteps), wantPending)
	}
}

// TestRunMigrate_RequiresSubcommand pins the CLI contract: bare
// `chronos migrate` is a usage error rather than a silent no-op.
func TestRunMigrate_RequiresSubcommand(t *testing.T) {
	if err := runMigrate(nil); err == nil {
		t.Fatal("expected usage error on empty args")
	}
}

// TestRunMigrate_UnknownSubcommandFails: typos must surface, not fall
// through to a default.
func TestRunMigrate_UnknownSubcommandFails(t *testing.T) {
	if err := runMigrate([]string{"redo"}); err == nil {
		t.Fatal("expected error on unknown subcommand")
	}
}

// TestRunMigrate_DownNotSupported: explicit error covers the
// forward-only contract the binary documents.
func TestRunMigrate_DownNotSupported(t *testing.T) {
	if err := runMigrate([]string{"down"}); err == nil {
		t.Fatal("expected error for unsupported down")
	}
}

// TestCurrentMigrationVersion_MatchesLadder is a sanity check on the
// status path: the binary's "currently applied" defaults to the
// highest declared step, so a fresh status output never says "0
// applied" when the auto-apply path has already run.
func TestCurrentMigrationVersion_MatchesLadder(t *testing.T) {
	if got, want := currentMigrationVersion(), chronosMigrations[len(chronosMigrations)-1].Version; got != want {
		t.Errorf("currentMigrationVersion() = %d, want %d", got, want)
	}
}
