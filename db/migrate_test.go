package db

import (
	"reflect"
	"testing"
)

func TestMigrationNames(t *testing.T) {
	want := []string{"migrations/000001_initial.up.sql", "migrations/000002_conversations.up.sql"}
	got, err := MigrationNames()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("migration names: got %v want %v", got, want)
	}
}
