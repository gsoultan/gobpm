// Package model is the storm schema: one plain Go struct per table, no tags.
//
// It replaces the GORM models in ../models, which carry their schema in struct
// tags a compiler cannot check. Here the type says what it can and a Schema
// method says the rest through field pointers, so a rename moves the constraint
// with the field instead of leaving a string behind that still compiles.
//
// Three conventions worth knowing before reading any of it:
//
//   - A non-pointer field is NOT NULL. Optional columns are pointers, which is
//     the opposite of the GORM models' habit of an empty string meaning absent.
//   - storm.JSON is a jsonb column carried as raw bytes. Everything the old
//     models stored through a GORM serializer — node graphs, token lists,
//     variable maps — is one of these, marshalled by the repository rather than
//     by a tag.
//   - Soft deletion is declared with t.SoftDelete(&x.DeletedAt), which compiles
//     the predicate into every read of that table and ANDs it ahead of the
//     caller's own. A call site can narrow what it sees and has no way to widen
//     it, so "remember the predicate" is not a property this codebase has to
//     hold. It was, briefly: storm before v0.8.0 left deleted_at an ordinary
//     column and every read spelled out DeletedAt.IsNull(), which is a rule
//     that only has to be forgotten once. DeletedAt stays declared per model
//     rather than hidden in an embedded Base, because which tables soft-delete
//     is a decision worth seeing.
//
// One consequence of that third point is worth its own paragraph, because it is
// the part that bites later. A marked row keeps its key, so a plain t.Unique on
// a soft-delete table is emitted as a partial index over the live rows — the
// value becomes available again once the row is removed. That is right for a
// connector key or an environment's port and wrong wherever something else
// references the row by id, or wherever reissuing the value would let a second
// subject inherit the first one's history: a version number, an idempotency
// key, a participant's username, a webhook token. Those declare
// t.UniqueAcrossDeleted, and each one says on the line above it what would
// happen if it did not.
package model
