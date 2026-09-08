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
	Origin    string
	Payload   string
	CreatedAt time.Time
}

func (b *BroadcastEvent) Schema(t *storm.Table) {
	t.PrimaryKey(&b.ID)
	t.Col(&b.ID).Identity()
	t.Col(&b.Origin).Size(64)
	t.Index(&b.Origin, &b.ID)
	t.Col(&b.CreatedAt).Index()
}
