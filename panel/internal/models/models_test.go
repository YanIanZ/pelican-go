package models

import (
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestUserModel(t *testing.T) {
	u := User{
		UUID:      "test-uuid",
		Email:     "test@example.com",
		Username:  "testuser",
		Password:  "hashed",
		Language:  "en",
		Timezone:  "UTC",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if u.TableName() != "users" {
		t.Errorf("expected table 'users', got '%s'", u.TableName())
	}
	if u.Email != "test@example.com" {
		t.Errorf("email mismatch")
	}
}

func TestServerModel(t *testing.T) {
	s := Server{
		UUID:      "server-uuid",
		UUIDShort: "srvr-shrt",
		Name:      "Test Server",
		Memory:    1024,
		Disk:      10240,
		CPU:       100,
		Status:    nil,
	}

	if s.TableName() != "servers" {
		t.Errorf("expected table 'servers', got '%s'", s.TableName())
	}
	if s.Status != nil {
		t.Error("new server should have nil status")
	}
}

func TestNodeModel(t *testing.T) {
	n := Node{
		UUID:   "node-uuid",
		Name:   "Test Node",
		FQDN:   "node.example.com",
		Scheme: "https",
	}

	if n.TableName() != "nodes" {
		t.Errorf("expected table 'nodes', got '%s'", n.TableName())
	}
	if n.MaintenanceMode != false {
		t.Error("new node should not be in maintenance mode")
	}
}

func TestAPIKeyTypes(t *testing.T) {
	if APIKeyTypeNone != 0 {
		t.Error("APIKeyTypeNone should be 0")
	}
	if APIKeyTypeAccount != 1 {
		t.Error("APIKeyTypeAccount should be 1")
	}
	if APIKeyTypeApplication != 2 {
		t.Error("APIKeyTypeApplication should be 2")
	}
}

func TestSoftDelete(t *testing.T) {
	s := Server{}
	if s.DeletedAt.Valid {
		t.Error("new server should not have deleted_at set")
	}

	var zeroTime gorm.DeletedAt
	if zeroTime.Valid {
		t.Error("zero DeletedAt should not be valid")
	}
}
