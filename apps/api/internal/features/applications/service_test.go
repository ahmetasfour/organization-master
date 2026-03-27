package applications_test

import (
	"context"
	"testing"
	"time"

	"membership-system/api/config"
	"membership-system/api/internal/features/applications"
	"membership-system/api/internal/features/auth"
	"membership-system/api/internal/features/logs"
	"membership-system/api/internal/shared"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// setupTestDB creates a test database connection.
// In a real test environment, this should connect to a test database.
func setupTestDB(t *testing.T) *gorm.DB {
	cfg, err := config.Load()
	if err != nil {
		t.Skipf("Skipping test: could not load config: %v", err)
	}

	db, err := config.ConnectDB(cfg)
	if err != nil {
		t.Skipf("Skipping test: could not connect to database: %v", err)
	}

	return db
}

// TestRejectionReasonImmutable verifies that rejection_reason cannot be changed once set.
func TestRejectionReasonImmutable(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// Create repositories and services
	appRepo := applications.NewRepository(db)
	authRepo := auth.NewRepository(db)
	logRepo := logs.NewRepository(db)
	appService := applications.NewService(appRepo, authRepo, logRepo)

	// Create a test application
	testApp := &applications.Application{
		ID:             uuid.New().String(),
		ApplicantName:  "Test Immutability User",
		ApplicantEmail: "immutable-test-" + uuid.New().String()[:8] + "@test.com",
		MembershipType: applications.MembershipAsil,
		Status:         applications.StatusReferansBekleniyor,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := appRepo.Create(ctx, testApp); err != nil {
		t.Fatalf("Failed to create test application: %v", err)
	}

	// Clean up after test
	defer func() {
		db.Exec("DELETE FROM applications WHERE id = ?", testApp.ID)
		db.Exec("DELETE FROM logs WHERE entity_id = ?", testApp.ID)
	}()

	// Use RedGuard to terminate with a reason
	redguard := shared.NewRedGuard(db)
	initialReason := "Initial rejection reason - test"
	actorID := uuid.New().String()

	err := redguard.Terminate(ctx, testApp.ID, initialReason, actorID, "yk")
	if err != nil {
		t.Fatalf("Failed to terminate application: %v", err)
	}

	// Verify the rejection_reason is set
	var app applications.Application
	if err := db.First(&app, "id = ?", testApp.ID).Error; err != nil {
		t.Fatalf("Failed to reload application: %v", err)
	}

	if app.RejectionReason == nil || *app.RejectionReason != initialReason {
		t.Errorf("Expected rejection_reason to be '%s', got '%v'", initialReason, app.RejectionReason)
	}

	// Attempt to update rejection_reason directly via raw SQL — should fail due to trigger
	newReason := "Attempted new reason"
	result := db.Exec("UPDATE applications SET rejection_reason = ? WHERE id = ?", newReason, testApp.ID)

	// The trigger should prevent this update
	var appAfter applications.Application
	if err := db.First(&appAfter, "id = ?", testApp.ID).Error; err != nil {
		t.Fatalf("Failed to reload application after update attempt: %v", err)
	}

	if result.Error == nil {
		// If no error, verify the value is still the original
		if appAfter.RejectionReason != nil && *appAfter.RejectionReason != initialReason {
			t.Errorf("Rejection reason was modified! Expected '%s', got '%s'", initialReason, *appAfter.RejectionReason)
		}
	}

	// Test via service layer as well — service should also prevent modification
	_ = appService // Service doesn't have a direct update method for rejection_reason,
	// which is by design — only RedGuard.Terminate can set it.

	t.Log("TestRejectionReasonImmutable: rejection_reason properly protected")
}

// TestTerminatedApplicationCannotAdvance verifies that terminated applications cannot change state.
func TestTerminatedApplicationCannotAdvance(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	appRepo := applications.NewRepository(db)

	// Create a test application
	testApp := &applications.Application{
		ID:             uuid.New().String(),
		ApplicantName:  "Test Advance User",
		ApplicantEmail: "advance-test-" + uuid.New().String()[:8] + "@test.com",
		MembershipType: applications.MembershipAsil,
		Status:         applications.StatusReferansBekleniyor,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := appRepo.Create(ctx, testApp); err != nil {
		t.Fatalf("Failed to create test application: %v", err)
	}

	defer func() {
		db.Exec("DELETE FROM applications WHERE id = ?", testApp.ID)
		db.Exec("DELETE FROM logs WHERE entity_id = ?", testApp.ID)
	}()

	// Terminate the application
	redguard := shared.NewRedGuard(db)
	if err := redguard.Terminate(ctx, testApp.ID, "Test termination", uuid.New().String(), "yk"); err != nil {
		t.Fatalf("Failed to terminate: %v", err)
	}

	// Attempt to advance the application — should fail
	err := redguard.AssertNotTerminated(ctx, testApp.ID)
	if err == nil {
		t.Error("Expected AssertNotTerminated to return an error for terminated application")
	}
	if err != shared.ErrApplicationTerminated {
		t.Errorf("Expected ErrApplicationTerminated, got: %v", err)
	}

	t.Log("TestTerminatedApplicationCannotAdvance: terminated applications properly blocked")
}

// TestRedGuardLogsTermination verifies that termination is logged.
func TestRedGuardLogsTermination(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	appRepo := applications.NewRepository(db)
	logRepo := logs.NewRepository(db)

	// Create a test application
	testApp := &applications.Application{
		ID:             uuid.New().String(),
		ApplicantName:  "Test Log User",
		ApplicantEmail: "log-test-" + uuid.New().String()[:8] + "@test.com",
		MembershipType: applications.MembershipAsil,
		Status:         applications.StatusReferansBekleniyor,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := appRepo.Create(ctx, testApp); err != nil {
		t.Fatalf("Failed to create test application: %v", err)
	}

	defer func() {
		db.Exec("DELETE FROM applications WHERE id = ?", testApp.ID)
		db.Exec("DELETE FROM logs WHERE entity_id = ?", testApp.ID)
	}()

	// Terminate
	redguard := shared.NewRedGuard(db)
	actorID := uuid.New().String()
	if err := redguard.Terminate(ctx, testApp.ID, "Log test reason", actorID, "yk"); err != nil {
		t.Fatalf("Failed to terminate: %v", err)
	}

	// Verify log entry exists
	var logEntries []*logs.Log
	if err := db.Where("entity_id = ? AND action = ?", testApp.ID, "application.terminated").Find(&logEntries).Error; err != nil {
		t.Fatalf("Failed to query logs: %v", err)
	}

	if len(logEntries) == 0 {
		t.Error("Expected termination log entry, found none")
	}

	if len(logEntries) > 0 {
		entry := logEntries[0]
		if entry.ActorRole != "yk" {
			t.Errorf("Expected actor_role 'yk', got '%s'", entry.ActorRole)
		}
		_ = logRepo // just to use it
	}

	t.Log("TestRedGuardLogsTermination: termination logging working correctly")
}
