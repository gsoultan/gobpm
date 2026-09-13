package model

import "github.com/gsoultan/storm"

// WorkflowGroupMembership puts a participant in a group.
//
// The pair is the primary key, so somebody cannot be added to a group twice and
// removing them once is enough.
type WorkflowGroupMembership struct {
	WorkflowUser  WorkflowUser
	WorkflowGroup WorkflowGroup
}

func (m *WorkflowGroupMembership) Schema(t *storm.Table) {
	t.PrimaryKey(&m.WorkflowUser, &m.WorkflowGroup)
	t.Col(&m.WorkflowUser).OnDelete(storm.Cascade)
	t.Col(&m.WorkflowGroup).OnDelete(storm.Cascade)
}
