package service

import (
	"context"
	"testing"

	"github.com/Rajneesh180/finance-backend/internal/domain"
	"github.com/Rajneesh180/finance-backend/internal/testutil"
	"github.com/google/uuid"
)

func newTestRecordService() *RecordService {
	repo := testutil.NewMockRecordRepo()
	dash := testutil.NewMockDashboardRepo()
	return NewRecordService(repo, dash)
}

func TestCreateRecord_Success(t *testing.T) {
	svc := newTestRecordService()
	userID := uuid.New()

	rec, err := svc.Create(context.Background(), userID, domain.CreateRecordRequest{
		Amount:   "1500.50",
		Type:     "income",
		Category: "salary",
		Date:     "2026-03-15",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if rec.UserID != userID {
		t.Errorf("got user_id %v, want %v", rec.UserID, userID)
	}
	if rec.Amount.String() != "1500.5" {
		t.Errorf("got amount %s, want 1500.5", rec.Amount.String())
	}
	if rec.Type != domain.RecordIncome {
		t.Errorf("got type %q, want %q", rec.Type, domain.RecordIncome)
	}
}

func TestCreateRecord_InvalidAmount(t *testing.T) {
	svc := newTestRecordService()
	_, err := svc.Create(context.Background(), uuid.New(), domain.CreateRecordRequest{
		Amount: "notanumber", Type: "income", Category: "test", Date: "2026-01-01",
	})
	if err != ErrInvalidAmount {
		t.Fatalf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestCreateRecord_NegativeAmount(t *testing.T) {
	svc := newTestRecordService()
	_, err := svc.Create(context.Background(), uuid.New(), domain.CreateRecordRequest{
		Amount: "-100", Type: "expense", Category: "test", Date: "2026-01-01",
	})
	if err != ErrInvalidAmount {
		t.Fatalf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestCreateRecord_InvalidDate(t *testing.T) {
	svc := newTestRecordService()
	_, err := svc.Create(context.Background(), uuid.New(), domain.CreateRecordRequest{
		Amount: "100", Type: "expense", Category: "test", Date: "not-a-date",
	})
	if err != ErrInvalidDate {
		t.Fatalf("expected ErrInvalidDate, got %v", err)
	}
}

func TestGetByID_OwnerAccess(t *testing.T) {
	svc := newTestRecordService()
	ctx := context.Background()
	userID := uuid.New()

	rec, _ := svc.Create(ctx, userID, domain.CreateRecordRequest{
		Amount: "200", Type: "expense", Category: "food", Date: "2026-03-10",
	})

	got, err := svc.GetByID(ctx, rec.ID, userID, false)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ID != rec.ID {
		t.Errorf("got id %v, want %v", got.ID, rec.ID)
	}
}

func TestGetByID_NotOwner(t *testing.T) {
	svc := newTestRecordService()
	ctx := context.Background()

	rec, _ := svc.Create(ctx, uuid.New(), domain.CreateRecordRequest{
		Amount: "200", Type: "expense", Category: "food", Date: "2026-03-10",
	})

	otherUser := uuid.New()
	_, err := svc.GetByID(ctx, rec.ID, otherUser, false)
	if err != ErrNotOwner {
		t.Fatalf("expected ErrNotOwner, got %v", err)
	}
}

func TestGetByID_AdminBypass(t *testing.T) {
	svc := newTestRecordService()
	ctx := context.Background()

	rec, _ := svc.Create(ctx, uuid.New(), domain.CreateRecordRequest{
		Amount: "200", Type: "expense", Category: "food", Date: "2026-03-10",
	})

	adminID := uuid.New()
	got, err := svc.GetByID(ctx, rec.ID, adminID, true)
	if err != nil {
		t.Fatalf("admin should bypass ownership: %v", err)
	}
	if got.ID != rec.ID {
		t.Errorf("got id %v, want %v", got.ID, rec.ID)
	}
}

func TestUpdateRecord_OwnerSuccess(t *testing.T) {
	svc := newTestRecordService()
	ctx := context.Background()
	userID := uuid.New()

	rec, _ := svc.Create(ctx, userID, domain.CreateRecordRequest{
		Amount: "100", Type: "expense", Category: "food", Date: "2026-03-10",
	})

	newAmt := "250.75"
	updated, err := svc.Update(ctx, rec.ID, userID, false, domain.UpdateRecordRequest{
		Amount: &newAmt,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Amount.String() != "250.75" {
		t.Errorf("got amount %s, want 250.75", updated.Amount.String())
	}
}

func TestDeleteRecord_NotOwner(t *testing.T) {
	svc := newTestRecordService()
	ctx := context.Background()

	rec, _ := svc.Create(ctx, uuid.New(), domain.CreateRecordRequest{
		Amount: "100", Type: "expense", Category: "food", Date: "2026-03-10",
	})

	err := svc.Delete(ctx, rec.ID, uuid.New(), false)
	if err != ErrNotOwner {
		t.Fatalf("expected ErrNotOwner, got %v", err)
	}
}

func TestDashboard(t *testing.T) {
	svc := newTestRecordService()
	summary, err := svc.Dashboard(context.Background(), nil, "", "")
	if err != nil {
		t.Fatalf("Dashboard: %v", err)
	}
	if summary.RecordCount != 5 {
		t.Errorf("got count %d, want 5", summary.RecordCount)
	}
}
