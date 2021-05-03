/*
Copyright 2020 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nopresource

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-template/apis/sample/v1alpha1"
)

// Unlike many Kubernetes projects Crossplane does not use third party testing
// libraries, per the common Go test review comments. Crossplane encourages the
// use of table driven unit tests. The tests of the crossplane-runtime project
// are representative of the testing style Crossplane encourages.
//
// https://github.com/golang/go/wiki/TestComments
// https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md#contributing-code

func TestObserve(t *testing.T) {
	type fields struct {
		service interface{}
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		o   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		// TODO: Add test cases.
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{service: tc.fields.service}
			got, err := e.Observe(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestReconcileLogic(t *testing.T) {

	c := []v1alpha1.ResourceConditionAfter{
		{Time: "10s", ConditionType: "Ready", ConditionStatus: "False"},
		{Time: "5s", ConditionType: "Ready", ConditionStatus: "False"},
		{Time: "7s", ConditionType: "Ready", ConditionStatus: "True"},
		{Time: "5s", ConditionType: "Synced", ConditionStatus: "False"},
		{Time: "10s", ConditionType: "Synced", ConditionStatus: "True"},
		{Time: "2s", ConditionType: "Ready", ConditionStatus: "False"},
	}

	cases := map[string]struct {
		reason            string
		resourcecondition []v1alpha1.ResourceConditionAfter
		elapsedtime       time.Duration
		want              []int
	}{
		"EmptyReconcileArray": {
			reason:            "Empty slice should be returned in case no conditions specified till given time elapsed.",
			resourcecondition: c,
			elapsedtime:       time.Duration(1*time.Second + 999*time.Millisecond),
			want:              []int{},
		},
		"SingleTypeReconcile": {
			reason:            "Slice with a single element should be returned when a single condition type has been specified.",
			resourcecondition: c,
			elapsedtime:       time.Duration(2 * time.Second),
			want:              []int{5},
		},
		"NormalReconcileBehaviour": {
			reason:            "Indexes with latest status of each condition type should be returned till given time elapsed.",
			resourcecondition: c,
			elapsedtime:       time.Duration(8 * time.Second),
			want:              []int{2, 3},
		},
		"LongTimeReconcileBehaviour": {
			reason:            "Indexes with last set status of each condition type should be returned till given time elapsed.",
			resourcecondition: c,
			elapsedtime:       time.Duration(50 * time.Second),
			want:              []int{0, 4},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := ReconcileLogic(tc.resourcecondition, tc.elapsedtime)
			if !sameIntSlice(tc.want, got) {
				t.Errorf("\n%s\nReconcileLogic(...): want []int %v, got []int %v\n", tc.reason, tc.want, got)
			}
		})
	}

}

// Matches the content of the slices. Order does not matter.
func sameIntSlice(x, y []int) bool {
	if len(x) != len(y) {
		return false
	}

	diff := make(map[int]int, len(x))
	for _, _x := range x {
		diff[_x]++
	}
	for _, _y := range y {
		if _, ok := diff[_y]; !ok {
			return false
		}
		diff[_y] -= 1
		if diff[_y] == 0 {
			delete(diff, _y)
		}
	}
	return len(diff) == 0
}
