package internal

import (
	"context"
	"testing"

	v1pb "buf.build/gen/go/bytebase/bytebase/protocolbuffers/go/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestMockBatchUpdateDatabasesRejectsInvalidConcreteParent(t *testing.T) {
	databaseName := "instances/instance-a/databases/database-a"
	client := &mockClient{
		databaseMap: map[string]*v1pb.Database{
			databaseName: {Name: databaseName},
		},
	}

	_, err := client.BatchUpdateDatabases(context.Background(), &v1pb.BatchUpdateDatabasesRequest{
		Parent: "instances/-",
		Requests: []*v1pb.UpdateDatabaseRequest{{
			Database:   &v1pb.Database{Name: databaseName, Project: "projects/project-a"},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"project"}},
		}},
	})
	if err == nil {
		t.Fatal("BatchUpdateDatabases accepted parent instances/-, want an invalid concrete parent error")
	}
}
