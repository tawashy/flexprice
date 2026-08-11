package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	baseMixin "github.com/flexprice/flexprice/ent/schema/mixin"
)

// IncomingWebhookEvent holds the schema for inbound webhook request audit logs.
type IncomingWebhookEvent struct {
	ent.Schema
}

func (IncomingWebhookEvent) Mixin() []ent.Mixin {
	return []ent.Mixin{
		baseMixin.BaseMixin{},
		baseMixin.EnvironmentMixin{},
	}
}

func (IncomingWebhookEvent) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			SchemaType(map[string]string{"postgres": "varchar(50)"}).
			Unique().
			Immutable(),
		field.String("provider").
			SchemaType(map[string]string{"postgres": "varchar(50)"}).
			NotEmpty().
			Immutable(),
		field.String("method").
			SchemaType(map[string]string{"postgres": "varchar(10)"}).
			NotEmpty().
			Immutable(),
		field.String("path").
			SchemaType(map[string]string{"postgres": "text"}).
			NotEmpty().
			Immutable(),
		field.String("request_id").
			SchemaType(map[string]string{"postgres": "varchar(100)"}).
			Optional().
			Immutable(),
		// provider_event_id is the provider's own event identifier (e.g. a Stripe
		// evt_* id). It is claimed by the provider handler after the body is parsed;
		// the partial unique index below makes that claim the redelivery-dedup
		// boundary. Deliberately mutable: the logging middleware inserts the row
		// before the body is parsed, so the id is attached later via
		// ClaimProviderEventID.
		field.String("provider_event_id").
			SchemaType(map[string]string{"postgres": "varchar(255)"}).
			Optional(),
		field.JSON("headers", map[string][]string{}).
			SchemaType(map[string]string{"postgres": "jsonb"}).
			Optional().
			Immutable(),
		field.Text("body").
			Optional().
			Immutable(),
	}
}

func (IncomingWebhookEvent) Edges() []ent.Edge {
	return nil
}

func (IncomingWebhookEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "environment_id", "provider", "created_at").
			StorageKey("idx_incoming_webhook_events_tenant_env_provider_created"),
		index.Fields("tenant_id", "environment_id", "created_at").
			StorageKey("idx_incoming_webhook_events_tenant_env_created"),
		index.Fields("request_id").
			StorageKey("idx_incoming_webhook_events_request_id"),
		// Redelivery dedup: at most one claimed row per (provider, provider_event_id).
		// Partial so audit-only rows (no event id) never collide.
		index.Fields("provider", "provider_event_id").
			Unique().
			StorageKey("idx_incoming_webhook_events_provider_event_unique").
			Annotations(entsql.IndexWhere("provider_event_id IS NOT NULL AND provider_event_id != ''")),
	}
}
