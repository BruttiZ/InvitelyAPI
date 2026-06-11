package tenants

import "context"

type Repository interface {
	Create(ctx context.Context, tenant Tenant) (Tenant, error)
	FindByID(ctx context.Context, id string) (Tenant, error)
}
