package internal

import "testing"

func TestInstanceResourceNames(t *testing.T) {
	tests := []struct {
		name       string
		parent     string
		resourceID string
	}{
		{
			name:       "instances/workspace-instance",
			resourceID: "workspace-instance",
		},
		{
			name:       "projects/sample-project/instances/project-instance",
			parent:     "projects/sample-project",
			resourceID: "project-instance",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := FormatInstanceName(test.parent, test.resourceID); got != test.name {
				t.Fatalf("FormatInstanceName(%q, %q) = %q, want %q", test.parent, test.resourceID, got, test.name)
			}

			parent, resourceID, err := GetInstanceParentAndID(test.name)
			if err != nil {
				t.Fatalf("GetInstanceParentAndID(%q) returned error: %v", test.name, err)
			}
			if parent != test.parent || resourceID != test.resourceID {
				t.Fatalf("GetInstanceParentAndID(%q) = (%q, %q), want (%q, %q)", test.name, parent, resourceID, test.parent, test.resourceID)
			}
		})
	}
}

func TestGetInstanceParentAndIDRejectsInvalidName(t *testing.T) {
	if _, _, err := GetInstanceParentAndID("projects/sample-project/instances"); err == nil {
		t.Fatal("GetInstanceParentAndID() error = nil, want invalid name error")
	}
}
