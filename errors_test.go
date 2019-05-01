package amznode_test

import (
	"fmt"
	"testing"

	"github.com/blacksails/amznode"
)

func TestNewErrNotFound(t *testing.T) {
	expectedID := 42
	expectedMsg := fmt.Sprintf("Could not find node with ID %d", expectedID)

	err := amznode.NewErrNotFound(42)

	if err.ID != expectedID {
		t.Errorf("expected id %d got %d", expectedID, err.ID)
	}
	if errMsg := err.Error(); errMsg != expectedMsg {
		t.Errorf("unexpected error message: expected '%s' got '%s'", expectedMsg, errMsg)
	}
}

func TestNewErrNameTaken(t *testing.T) {
	expectedName := "test"
	expectedParentID := 42
	expectedMsg := fmt.Sprintf(
		"the name '%s' has already been taken under the parent with id #%d",
		expectedName, expectedParentID,
	)

	err := amznode.NewErrNameTaken("test", 42)

	if err.Name != expectedName {
		t.Errorf("expected name '%s' got '%s'", expectedName, err.Name)
	}
	if err.ParentID != expectedParentID {
		t.Errorf("expected id %d got %d", expectedParentID, err.ParentID)
	}
	if errMsg := err.Error(); errMsg != expectedMsg {
		t.Errorf("unexpected error message: expected '%s' got '%s'", expectedMsg, errMsg)
	}
}

func TestNewNodeIsDecendant(t *testing.T) {
	expectedID := 0
	expectedDecendantID := 1
	expectedMsg := fmt.Sprintf(
		"the node with id %d is a decendant of the node with id %d",
		expectedDecendantID, expectedID,
	)

	err := amznode.NewErrNodeIsDecendant(0, 1)

	if err.ID != expectedID {
		t.Errorf("expected id %d got %d", expectedID, err.ID)
	}
	if err.DecendantID != expectedDecendantID {
		t.Errorf("expected id %d got %d", expectedDecendantID, err.DecendantID)
	}
	if errMsg := err.Error(); errMsg != expectedMsg {
		t.Errorf("unexpected error message: expected '%s' got '%s'", expectedMsg, errMsg)
	}
}
