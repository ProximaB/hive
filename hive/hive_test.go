package hive

import (
	"fmt"
	"log"
	"testing"

	"github.com/pkg/errors"
)

type generic struct{}

// Run runs a generic job
func (g generic) Run(job Job, do DoFunc) (interface{}, error) {
	if job.String() == "first" {
		return do(NewJob("generic", "second")), nil
	} else if job.String() == "second" {
		return do(NewJob("generic", "last")), nil
	} else if job.String() == "fail" {
		return nil, errors.New("error!")
	}

	return job.String(), nil
}

func (g generic) OnStart() error {
	return nil
}

func TestHiveJob(t *testing.T) {
	h := New()

	h.Handle("generic", generic{})

	r := h.Do(h.Job("generic", "first"))

	if r.UUID() == "" {
		t.Error("result ID is empty")
	}

	res, err := r.Then()
	if err != nil {
		log.Fatal(err)
	}

	if res.(string) != "last" {
		t.Error("generic job failed, expected 'last', got", res.(string))
	}
}

type input struct {
	First, Second int
}

type math struct{}

// Run runs a math job
func (g math) Run(job Job, do DoFunc) (interface{}, error) {
	in := job.Data().(input)

	return in.First + in.Second, nil
}

func (g math) OnStart() error {
	return nil
}

func TestHiveJobHelperFunc(t *testing.T) {
	h := New()

	doMath := h.Handle("math", math{})

	for i := 1; i < 10; i++ {
		answer := i + i*3

		equals, _ := doMath(input{i, i * 3}).ThenInt()
		if equals != answer {
			t.Error("failed to get math right, expected", answer, "got", equals)
		}
	}
}

func TestHiveResultDiscard(t *testing.T) {
	h := New()

	h.Handle("generic", generic{})

	r := h.Do(h.Job("generic", "first"))

	// basically just making sure that it doesn't hold up the line
	r.Discard()
}

func TestHiveResultThenDo(t *testing.T) {
	h := New()

	h.Handle("generic", generic{})

	wait := make(chan bool)

	h.Do(h.Job("generic", "first")).ThenDo(func(res interface{}, err error) {
		if err != nil {
			t.Error(errors.Wrap(err, "did not expect error"))
			wait <- false
		}

		if res.(string) != "last" {
			t.Error(fmt.Errorf("expected 'last', got %s", res.(string)))
		}

		wait <- true
	})

	h.Do(h.Job("generic", "fail")).ThenDo(func(res interface{}, err error) {
		if err == nil {
			t.Error(errors.New("expected error, did not get one"))
			wait <- false
		}

		wait <- true
	})

	// poor man's async testing
	<-wait
	<-wait
}
