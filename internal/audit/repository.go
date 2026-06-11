package audit

import "context"

type Repository interface {
	Create(ctx context.Context, entry Entry) (Entry, error)
	ListByTenant(ctx context.Context, tenantID string) ([]Entry, error)
}
