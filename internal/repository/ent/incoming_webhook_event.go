package ent

import (
	"context"

	"github.com/flexprice/flexprice/ent"
	"github.com/flexprice/flexprice/ent/incomingwebhookevent"
	domainIncomingWebhookEvent "github.com/flexprice/flexprice/internal/domain/incomingwebhookevent"
	ierr "github.com/flexprice/flexprice/internal/errors"
	"github.com/flexprice/flexprice/internal/postgres"
)

type incomingWebhookEventRepository struct {
	client postgres.IClient
}

// NewIncomingWebhookEventRepository creates a new Ent-backed incoming webhook event repository.
func NewIncomingWebhookEventRepository(client postgres.IClient) domainIncomingWebhookEvent.Repository {
	return &incomingWebhookEventRepository{client: client}
}

func (r *incomingWebhookEventRepository) Create(ctx context.Context, event *domainIncomingWebhookEvent.IncomingWebhookEvent) error {
	client := r.client.Writer(ctx)

	create := client.IncomingWebhookEvent.Create().
		SetID(event.ID).
		SetTenantID(event.TenantID).
		SetEnvironmentID(event.EnvironmentID).
		SetProvider(event.Provider).
		SetMethod(event.Method).
		SetPath(event.Path).
		SetRequestID(event.RequestID).
		SetHeaders(event.Headers).
		SetBody(event.Body)
	if event.ProviderEventID != "" {
		create.SetProviderEventID(event.ProviderEventID)
	}

	_, err := create.Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) && event.ProviderEventID != "" {
			return ierr.WithError(err).
				WithHint("Webhook event with this provider event id was already recorded").
				WithReportableDetails(map[string]any{
					"provider":          event.Provider,
					"provider_event_id": event.ProviderEventID,
				}).
				Mark(ierr.ErrAlreadyExists)
		}
		return ierr.WithError(err).
			WithHint("Failed to persist incoming webhook event log").
			WithReportableDetails(map[string]any{
				"provider":       event.Provider,
				"tenant_id":      event.TenantID,
				"environment_id": event.EnvironmentID,
			}).
			Mark(ierr.ErrDatabase)
	}
	return nil
}

func (r *incomingWebhookEventRepository) ClaimProviderEventID(ctx context.Context, eventID string, providerEventID string) error {
	if providerEventID == "" {
		return ierr.NewError("provider_event_id is required").
			WithHint("Cannot claim an empty provider event id").
			Mark(ierr.ErrValidation)
	}
	client := r.client.Writer(ctx)

	err := client.IncomingWebhookEvent.UpdateOneID(eventID).
		SetProviderEventID(providerEventID).
		Exec(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return ierr.WithError(err).
				WithHint("This provider event was already processed (duplicate delivery)").
				WithReportableDetails(map[string]any{
					"event_id":          eventID,
					"provider_event_id": providerEventID,
				}).
				Mark(ierr.ErrAlreadyExists)
		}
		return ierr.WithError(err).
			WithHint("Failed to claim provider event id").
			WithReportableDetails(map[string]any{
				"event_id":          eventID,
				"provider_event_id": providerEventID,
			}).
			Mark(ierr.ErrDatabase)
	}
	return nil
}

func (r *incomingWebhookEventRepository) ExistsByProviderEventID(ctx context.Context, provider string, providerEventID string) (bool, error) {
	client := r.client.Reader(ctx)

	exists, err := client.IncomingWebhookEvent.Query().
		Where(
			incomingwebhookevent.Provider(provider),
			incomingwebhookevent.ProviderEventID(providerEventID),
		).
		Exist(ctx)
	if err != nil {
		return false, ierr.WithError(err).
			WithHint("Failed to check provider event id").
			WithReportableDetails(map[string]any{
				"provider":          provider,
				"provider_event_id": providerEventID,
			}).
			Mark(ierr.ErrDatabase)
	}
	return exists, nil
}
