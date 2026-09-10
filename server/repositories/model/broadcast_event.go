package model

import (
	"time"

	"github.com/gsoultan/storm"
)

// BroadcastEvent is one entry on the fan-out bus that carries live UI updates
// between replicas.
//
// Without it a browser only sees events produced by the replica it happens to
// be connected to. Readers poll for rows newer than the last id they saw, which
// is why the id is a sequence rather than a uuid: it orders, and a uuid does
// not.
type BroadcastEvent struct {
	ID int64
	// Origin is the replica that produced it, so a reader can skip its own.
	Origin  string
	Payload string

	// OrganizationID and EnvironmentID are who the event is for.
	//
	// They travel with the payload because the replica that delivers it has no
	// context to recover them from: the request that produced it happened on
	// another machine. Nullable so rows written before the columns existed
	// still read; those are dropped on delivery rather than broadcast, because
	// an event whose audience is unknown has no safe audience.
	// Named for the relation, not the column: storm appends ID, so a field
	// called OrganizationID would produce organization_id_id.
	Organization *Organization
	Environment  *Environment

	CreatedAt time.Time
}

func (b *BroadcastEvent) Schema(t *storm.Table) {
	t.PrimaryKey(&b.ID)
	t.Col(&b.ID).Identity()
	t.Col(&b.Origin).Size(64)
	t.Index(&b.Origin, &b.ID)
	t.Col(&b.CreatedAt).Index()
	// The bus is pruned, not cascaded: an event outlives nothing, and deleting
	// an organization must not fail on rows that are about to be swept anyway.
	t.Col(&b.Organization).OnDelete(storm.SetNull).Index()
	t.Col(&b.Environment).OnDelete(storm.SetNull)
}
