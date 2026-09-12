package entities

import "time"

// OCEL 2.0 — the object-centric event log this engine's audit trail can be read
// as, so that a process instance's history can be mined by the tools that exist
// rather than only viewed in this application.
//
// The audit trail is already a complete record of what happened: one row per
// state change, with the node, the instance and the time. What it was not was
// *portable* — reading it meant using this UI. OCEL 2.0 is the current standard
// for exchanging object-centric event data, supported by ProM, pm4py and the
// commercial mining tools, and it is a strictly richer target than XES because
// one event can relate to several objects at once.
//
// The field names below are the JSON exchange format's, not Go's conventions;
// a reader keys off them exactly, so they are not negotiable.

// OCELLog is a complete object-centric event log.
type OCELLog struct {
	ObjectTypes []OCELType   `json:"objectTypes"`
	EventTypes  []OCELType   `json:"eventTypes"`
	Objects     []OCELObject `json:"objects"`
	Events      []OCELEvent  `json:"events"`
}

// OCELType declares a type and the attributes its members may carry.
type OCELType struct {
	Name       string              `json:"name"`
	Attributes []OCELAttributeDecl `json:"attributes"`
}

// OCELAttributeDecl is one attribute a type may carry. The type name is the
// standard's own vocabulary — "string", "time", "integer", "float", "boolean".
type OCELAttributeDecl struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// OCELObject is a thing the process acts on or with. Here that is a process
// instance or the person who acted.
type OCELObject struct {
	ID            string             `json:"id"`
	Type          string             `json:"type"`
	Attributes    []OCELObjectAttr   `json:"attributes"`
	Relationships []OCELRelationship `json:"relationships,omitempty"`
}

// OCELObjectAttr is an object attribute at a point in time. OCEL 2.0's object
// attributes are time-stamped because an object's properties change over its
// life; an attribute that never changes carries the time it was first known.
type OCELObjectAttr struct {
	Name  string    `json:"name"`
	Time  time.Time `json:"time"`
	Value string    `json:"value"`
}

// OCELEvent is one thing that happened.
type OCELEvent struct {
	ID            string             `json:"id"`
	Type          string             `json:"type"`
	Time          time.Time          `json:"time"`
	Attributes    []OCELEventAttr    `json:"attributes"`
	Relationships []OCELRelationship `json:"relationships,omitempty"`
}

// OCELEventAttr is an event attribute. Unlike an object attribute it carries no
// time of its own: it is true as of the event.
type OCELEventAttr struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// OCELRelationship names the role an object played.
//
// The qualifier is the part that makes an object-centric log worth more than a
// flat one: "who approved this" and "what was approved" are both relationships
// to the same event, distinguished only by this string.
type OCELRelationship struct {
	ObjectID  string `json:"objectId"`
	Qualifier string `json:"qualifier"`
}

// The object and relationship vocabulary this engine emits. They are constants
// because a mining tool's configuration refers to them by name, so changing one
// silently invalidates whatever somebody built on top of the export.
const (
	OCELObjectProcessInstance = "ProcessInstance"
	OCELObjectDefinition      = "ProcessDefinition"

	OCELQualifierInstance = "instance"
	OCELQualifierDefines  = "definition"
)

// OCELOptions controls what an export includes.
type OCELOptions struct {
	// IncludeVariables adds each audit entry's data to the exported events.
	//
	// It is off by default and it has to stay that way. An audit entry's data
	// map is the instance's process variables — an amount, an applicant's name,
	// an approval decision — so an export that includes them by default turns a
	// convenience endpoint into a way to take every business fact in a project
	// out of the building in one request. Mining a control-flow model needs the
	// activity, the case and the time, and none of those are in there.
	IncludeVariables bool
}
