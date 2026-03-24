package contracts

// VariableHistoryService composes VariableHistoryWriter and VariableHistoryReader
// into the full variable history contract. Inject this into the engine to enable
// SOX/GDPR-grade audit of process variable changes at every execution step.
type VariableHistoryService interface {
	VariableHistoryWriter
	VariableHistoryReader
}
