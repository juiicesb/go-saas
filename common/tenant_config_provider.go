package common

import "context"

type TenantConfigProvider interface {
	// Get tenant config
	Get(ctx context.Context) (TenantConfig, context.Context, error)
}

type DefaultTenantConfigProvider struct {
	tr TenantResolver
	ts TenantStore
}

func NewDefaultTenantConfigProvider(tr TenantResolver, ts TenantStore) TenantConfigProvider {
	return &DefaultTenantConfigProvider{
		tr: tr,
		ts: ts,
	}
}

func (d *DefaultTenantConfigProvider) Get(ctx context.Context) (TenantConfig, context.Context, error) {
	rr, ctx, err := d.tr.Resolve(ctx)
	if err != nil {
		return TenantConfig{}, ctx, err
	}
	if rr.TenantIdOrName != "" {
		//tenant side
		//get config from tenant store
		cfg, err := d.ts.GetByNameOrId(ctx, rr.TenantIdOrName)
		if err != nil {
			return TenantConfig{}, ctx, err
		}
		return *cfg, ctx, nil
		//check error
	}
	return TenantConfig{}, ctx, nil

}
