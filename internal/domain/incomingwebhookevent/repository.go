package incomingwebhookevent

import "context"

// Repository defines the persistence interface for incoming webhook event audit logs.
type Repository interface {
	Create(ctx context.Context, event *IncomingWebhookEvent) error
	// ClaimProviderEventID attaches the provider's event id to a previously logged
	// event row. The partial unique index on (provider, provider_event_id) makes the
	// claim the redelivery-dedup boundary: a duplicate delivery fails with
	// ErrAlreadyExists, which callers treat as "already processed". Run it inside the
	// same transaction as the event's side effects for exactly-once semantics.
	ClaimProviderEventID(ctx context.Context, eventID string, providerEventID string) error
	// ExistsByProviderEventID reports whether a claimed row exists for the provider's
	// event id (read-only pre-check; the claim is the authoritative guard).
	ExistsByProviderEventID(ctx context.Context, provider string, providerEventID string) (bool, error)
}
