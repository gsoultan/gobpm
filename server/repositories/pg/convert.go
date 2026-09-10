package pg

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/gsoultan/storm/runtime"
)

// jsonOf encodes a map for a jsonb column.
//
// Nil rather than "null" for an absent map: a NULL column and a column holding
// the four bytes `null` read back the same in Go and differently in SQL, and
// only one of them is what "there was nothing here" means.
func jsonOf(m map[string]any) (runtime.JSON, error) {
	if len(m) == 0 {
		return nil, nil
	}
	encoded, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

// mapOf decodes a jsonb column, treating absent as an absent map rather than an
// empty one — the callers that check `len(vars) == 0` want the same answer for
// both, and the ones that range over it want neither to panic.
func mapOf(raw runtime.JSON) (map[string]any, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// uuidsToRaw converts a list of ids for an IN predicate.
//
// storm's generated predicates take [16]byte, which uuid.UUID is — but a
// []uuid.UUID is not a [][16]byte, and Go will not convert the slice for us.
func uuidsToRaw(ids []uuid.UUID) [][16]byte {
	out := make([][16]byte, len(ids))
	for i, id := range ids {
		out[i] = id
	}
	return out
}
