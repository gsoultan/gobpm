package model

// MultiInstanceState is how far a multi-instance node has got.
//
// Not a table: it lives inside ProcessInstance.MultiInstance as jsonb, keyed by
// node id. Its own column rather than a business variable so a process cannot
// overwrite the engine's own counters by naming a variable the same thing.
type MultiInstanceState struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
}
