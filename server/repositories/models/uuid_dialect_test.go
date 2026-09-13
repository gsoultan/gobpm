package models

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// The column type an identifier gets.
//
// This is the decision every table's primary key depends on, and getting it
// wrong is not subtle: AutoMigrate fails on the first table and the server
// cannot start at all. MySQL was once in exactly that state — offered in the
// config and the setup wizard, and unable to create a single table — because
// every test ran on SQLite. It is not offered any more, and this is what is
// left of the check that found it.
//
// SQLite is still here because the test harness uses it; the product does not.
func TestAnIdentifierGetsAColumnTypeThatExists(t *testing.T) {
	for _, want := range []struct {
		dialect gorm.Dialector
		column  string
		why     string
	}{
		{postgres.New(postgres.Config{}), "uuid", "PostgreSQL has a native uuid type, and changing this would ALTER TABLE every existing deployment on its next boot"},
		{sqlite.Open(":memory:"), "uuid", "the test harness runs on SQLite, and its databases were created with this type name"},
	} {
		db := &gorm.DB{Config: &gorm.Config{Dialector: want.dialect}}
		if got := NilUUID.GormDBDataType(db, nil); got != want.column {
			t.Errorf("%s gets %q, want %q — %s", want.dialect.Name(), got, want.column, want.why)
		}
	}
}

// With no database to ask, there is still a type to name.
func TestTheColumnTypeSurvivesHavingNoDialector(t *testing.T) {
	if got := NilUUID.GormDBDataType(nil, nil); got == "" {
		t.Error("a nil db gave no column type at all")
	}
	if got := NilUUID.GormDBDataType(&gorm.DB{Config: &gorm.Config{}}, nil); got == "" {
		t.Error("a db with no dialector gave no column type at all")
	}
}
