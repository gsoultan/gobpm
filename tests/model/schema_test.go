package model_test

import (
	"testing"

	"github.com/gsoultan/metis/server/repositories/model"
	"github.com/gsoultan/storm"
	"github.com/gsoultan/storm/compile/pgddl"
)

// The real model layer builds, and its DDL is what we expect.
func TestTheModelLayerBuilds(t *testing.T) {
	s, err := storm.Build(model.All()...)
	if err != nil {
		t.Fatalf("storm.Build refused the model layer:\n%v", err)
	}
	t.Logf("DDL:\n%s", pgddl.Create(s))
}
