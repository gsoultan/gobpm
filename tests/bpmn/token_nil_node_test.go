package bpmn_test

import (
	"testing"

	"github.com/gsoultan/metis/server/domains/entities"
)

// A token whose node is absent from the loaded definition used to reach
// RemoveTokenByNode as a nil, where the comparison dereferenced it. That did not
// panic and did not return — the goroutine spun forever holding its transaction,
// so the request never answered and nothing was logged.
//
// The bound is what makes this a regression test rather than a hang: before the
// fix these never finished at all.
func TestRemoveTokenByNilNodeReturns(t *testing.T) {
	instance := &entities.ProcessInstance{
		Tokens: []entities.Token{{Node: &entities.Node{ID: "task1"}}},
	}

	instance.RemoveTokenByNode(nil)
	if len(instance.Tokens) != 1 {
		t.Fatalf("a nil node should remove nothing, got %d tokens", len(instance.Tokens))
	}

	instance.RemoveTokenByIteration(nil, "0")
	if len(instance.Tokens) != 1 {
		t.Fatalf("a nil node should remove no iteration, got %d tokens", len(instance.Tokens))
	}

	if got := instance.GetTokensByNode(nil); got != nil {
		t.Fatalf("a nil node holds no tokens, got %d", len(got))
	}
}

// The engine knows the ID it is advancing past even when the definition has
// stopped describing that node, so the token still comes off.
func TestRemoveTokenByNodeIDClearsAVanishedNode(t *testing.T) {
	instance := &entities.ProcessInstance{
		Tokens: []entities.Token{
			{Node: &entities.Node{ID: "task1"}},
			{Node: &entities.Node{ID: "task2"}},
		},
	}

	instance.RemoveTokenByNodeID("task1")

	if len(instance.Tokens) != 1 {
		t.Fatalf("expected one token left, got %d", len(instance.Tokens))
	}
	if instance.Tokens[0].Node.ID != "task2" {
		t.Fatalf("removed the wrong token, %q is left", instance.Tokens[0].Node.ID)
	}
}
