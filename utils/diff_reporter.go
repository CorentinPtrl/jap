package utils

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"strings"
)

type DiffReporter struct {
	path   cmp.Path
	diffs  []string
	fields []string
}

func (r *DiffReporter) PushStep(ps cmp.PathStep) {
	r.path = append(r.path, ps)
}

func (r *DiffReporter) Report(rs cmp.Result) {
	if !rs.Equal() {
		vx, vy := r.path.Last().Values()
		if vy.IsValid() {
			r.fields = append(r.fields, r.path.Index(1).String())
			r.diffs = append(r.diffs, fmt.Sprintf("%#v:\n\t-: %+v\n\t+: %+v\n", r.path, vx, vy))
		}
	}
}

func (r *DiffReporter) PopStep() {
	r.path = r.path[:len(r.path)-1]
}

func (r DiffReporter) Fields() []string {
	return r.fields
}

func (r *DiffReporter) String() string {
	return strings.Join(r.fields, "\n")
}
