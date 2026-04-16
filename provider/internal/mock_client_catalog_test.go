package internal

import (
	"context"
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
)

// Verify that a catalog written via UpdateDatabaseCatalog is readable via
// GetDatabaseCatalog using the database name the real client would pass in.
// The real client calls GetDatabaseCatalog(ctx, databaseName) where
// databaseName does NOT include the "/catalog" suffix; the suffix is added
// inside the catalog proto's Name field by convertToV1DatabaseCatalog.
func TestMockDatabaseCatalog_WriteThenReadRoundTrip(t *testing.T) {
	c := &mockClient{
		databaseCatalogMap: map[string]*v1pb.DatabaseCatalog{},
	}
	databaseName := "instances/test-inst/databases/test-db"
	patch := &v1pb.DatabaseCatalog{
		Name: databaseName + DatabaseCatalogNameSuffix,
	}
	if _, err := c.UpdateDatabaseCatalog(context.Background(), patch); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, err := c.GetDatabaseCatalog(context.Background(), databaseName)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != patch.Name {
		t.Errorf("catalog name mismatch: got %q want %q", got.Name, patch.Name)
	}
}
