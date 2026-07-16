package vector

import (
	"context"
	"fmt"

	"github.com/weaviate/weaviate/entities/models"
)

// CreateTenant creates a new tenant for native multi-tenancy
func (s *Store) CreateTenant(ctx context.Context, tenantName string) error {
	// Use the default class for tenant creation
	className, err := s.getClassName(ctx, "")
	if err != nil {
		return err
	}

	// Create tenant using the Weaviate Go client
	tenant := models.Tenant{
		Name: tenantName,
	}

	err = s.client.Schema().TenantsCreator().
		WithClassName(className).
		WithTenants(tenant).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to create tenant %s: %w", tenantName, err)
	}

	s.logger.Info(ctx, "Tenant created successfully", map[string]interface{}{
		"tenantName": tenantName,
		"className":  className,
	})
	return nil
}

// DeleteTenant deletes a tenant for native multi-tenancy
func (s *Store) DeleteTenant(ctx context.Context, tenantName string) error {
	// Use the default class for tenant deletion
	className, err := s.getClassName(ctx, "")
	if err != nil {
		return err
	}

	err = s.client.Schema().TenantsDeleter().
		WithClassName(className).
		WithTenants(tenantName).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete tenant %s: %w", tenantName, err)
	}

	s.logger.Info(ctx, "Tenant deleted successfully", map[string]interface{}{
		"tenantName": tenantName,
		"className":  className,
	})
	return nil
}

// ListTenants lists all tenants for native multi-tenancy
func (s *Store) ListTenants(ctx context.Context) ([]string, error) {
	// Use the default class for tenant listing
	className, err := s.getClassName(ctx, "")
	if err != nil {
		return nil, err
	}

	tenants, err := s.client.Schema().TenantsGetter().
		WithClassName(className).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	var tenantNames []string
	for _, tenant := range tenants {
		tenantNames = append(tenantNames, tenant.Name)
	}

	s.logger.Info(ctx, "Tenants listed successfully", map[string]interface{}{
		"className": className,
		"count":     len(tenantNames),
	})
	return tenantNames, nil
}
