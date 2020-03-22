package hive

import (
	"testing"

	"github.com/pkg/errors"
)

func TestHiveJobGroup(t *testing.T) {
	h := New()

	doMath := h.Handle("math", math{})

	grp := NewGroup()
	grp.Add(doMath(input{5, 6}))
	grp.Add(doMath(input{7, 8}))
	grp.Add(doMath(input{9, 10}))

	if err := grp.Wait(); err != nil {
		t.Error(errors.Wrap(err, "failed to grp.Wait"))
	}
}

type groupWork struct{}

// Run runs a groupWork job
func (g groupWork) Run(job Job, run RunFunc) (interface{}, error) {
	grp := NewGroup()

	grp.Add(run(NewJob("generic", "first")))
	grp.Add(run(NewJob("generic", "group work")))
	grp.Add(run(NewJob("generic", "group work")))
	grp.Add(run(NewJob("generic", "group work")))
	grp.Add(run(NewJob("generic", "group work")))

	return grp, nil
}

func TestHiveChainedGroup(t *testing.T) {
	h := New()

	h.Handle("generic", generic{})
	doGrp := h.Handle("group", groupWork{})

	if _, err := doGrp(nil).Then(); err != nil {
		t.Error(errors.Wrap(err, "failed to doGrp"))
	}
}