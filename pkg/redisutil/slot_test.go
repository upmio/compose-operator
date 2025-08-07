/*
Copyright 2025 The Compose Operator Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

SPDX-License-Identifier: Apache-2.0
*/

package redisutil

import (
	"reflect"
	"sort"
	"testing"
)

func TestRemoveSlots(t *testing.T) {
	type args struct {
		slots        []Slot
		removedSlots []Slot
	}
	tests := []struct {
		name string
		args args
		want []Slot
	}{
		{
			name: "Case 1",
			args: args{
				slots:        []Slot{2, 3, 4, 5, 6, 7, 8, 9, 10},
				removedSlots: []Slot{2, 10},
			},
			want: []Slot{3, 4, 5, 6, 7, 8, 9},
		},
		{
			name: "Case 2",
			args: args{
				slots:        []Slot{2, 5},
				removedSlots: []Slot{2, 2, 3},
			},
			want: []Slot{5},
		},
		{
			name: "Case 3",
			args: args{
				slots:        []Slot{0, 1, 3, 4},
				removedSlots: []Slot{0, 1, 3, 4},
			},
			want: []Slot{},
		},
		{
			name: "Case 4",
			args: args{
				slots:        []Slot{},
				removedSlots: []Slot{2, 10},
			},
			want: []Slot{},
		},
		{
			name: "Case 5",
			args: args{
				slots:        []Slot{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				removedSlots: []Slot{5},
			},
			want: []Slot{0, 1, 2, 3, 4, 6, 7, 8, 9, 10},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveSlots(tt.args.slots, tt.args.removedSlots); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveSlots() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveSlot(t *testing.T) {
	type args struct {
		slots       []Slot
		removedSlot Slot
	}
	tests := []struct {
		name string
		args args
		want []Slot
	}{
		{
			name: "Case 1",
			args: args{
				slots:       []Slot{2, 3, 4, 5, 6, 7, 8, 9, 10},
				removedSlot: 2,
			},
			want: []Slot{3, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			name: "Case 2",
			args: args{
				slots:       []Slot{2, 5},
				removedSlot: 2,
			},
			want: []Slot{5},
		},
		{
			name: "Case 3",
			args: args{
				slots:       []Slot{0, 1, 3, 4},
				removedSlot: 3,
			},
			want: []Slot{0, 1, 4},
		},
		{
			name: "Case 4",
			args: args{
				slots:       []Slot{},
				removedSlot: 2,
			},
			want: []Slot{},
		},
		{
			name: "Case 5",
			args: args{
				slots:       []Slot{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				removedSlot: 5,
			},
			want: []Slot{0, 1, 2, 3, 4, 6, 7, 8, 9, 10},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveSlot(tt.args.slots, tt.args.removedSlot); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveSlot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSlotString(t *testing.T) {
	slot := Slot(42)
	expected := "42"
	if slot.String() != expected {
		t.Errorf("expected %s, got %s", expected, slot.String())
	}
}

func TestSlotSliceString(t *testing.T) {
	slots := SlotSlice{1, 2, 3}
	expected := "[1-3]"
	if slots.String() != expected {
		t.Errorf("expected %s, got %s", expected, slots.String())
	}
}

func TestDecodeSlot(t *testing.T) {
	input := "42"
	expected := Slot(42)
	slot, err := DecodeSlot(input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if slot != expected {
		t.Errorf("expected %v, got %v", expected, slot)
	}
}

func TestSlotRangeString(t *testing.T) {
	slotRange := SlotRange{Min: 1, Max: 10}
	expected := "1-10"
	if slotRange.String() != expected {
		t.Errorf("expected %s, got %s", expected, slotRange.String())
	}
}

func TestSlotRangeTotal(t *testing.T) {
	slotRange := SlotRange{Min: 1, Max: 10}
	expected := 10
	if slotRange.Total() != expected {
		t.Errorf("expected %d, got %d", expected, slotRange.Total())
	}
}

func TestImportingSlotString(t *testing.T) {
	importingSlot := ImportingSlot{SlotID: 42, FromNodeID: "node1"}
	expected := "42-<-node1"
	if importingSlot.String() != expected {
		t.Errorf("expected %s, got %s", expected, importingSlot.String())
	}
}

func TestMigratingSlotString(t *testing.T) {
	migratingSlot := MigratingSlot{SlotID: 42, ToNodeID: "node2"}
	expected := "42->-node2"
	if migratingSlot.String() != expected {
		t.Errorf("expected %s, got %s", expected, migratingSlot.String())
	}
}

func TestDecodeSlotRange(t *testing.T) {
	testCases := []struct {
		input             string
		expectedSlots     []Slot
		expectedImporting *ImportingSlot
		expectedMigrating *MigratingSlot
		expectError       bool
	}{
		{"42", []Slot{42}, nil, nil, false},
		{"42-45", []Slot{42, 43, 44, 45}, nil, nil, false},
		{"[42-<-node1]", []Slot{}, &ImportingSlot{SlotID: 42, FromNodeID: "node1"}, nil, false},
		{"[42->-node2]", []Slot{}, nil, &MigratingSlot{SlotID: 42, ToNodeID: "node2"}, false},
		{"invalid", nil, nil, nil, true},
	}

	for _, tc := range testCases {
		slots, importing, migrating, err := DecodeSlotRange(tc.input)
		if tc.expectError {
			if err == nil {
				t.Errorf("expected error for input %s", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for input %s: %v", tc.input, err)
			}
			if !reflect.DeepEqual(slots, tc.expectedSlots) {
				t.Errorf("expected slots %v, got %v for input %s", tc.expectedSlots, slots, tc.input)
			}
			if !reflect.DeepEqual(importing, tc.expectedImporting) {
				t.Errorf("expected importing slot %v, got %v for input %s", tc.expectedImporting, importing, tc.input)
			}
			if !reflect.DeepEqual(migrating, tc.expectedMigrating) {
				t.Errorf("expected migrating slot %v, got %v for input %s", tc.expectedMigrating, migrating, tc.input)
			}
		}
	}
}

func TestSlotRangesFromSlots(t *testing.T) {
	slots := []Slot{1, 2, 3, 5, 6, 7, 9}
	expected := []SlotRange{
		{Min: 1, Max: 3},
		{Min: 5, Max: 7},
		{Min: 9, Max: 9},
	}
	ranges := SlotRangesFromSlots(slots)
	if !reflect.DeepEqual(ranges, expected) {
		t.Errorf("expected %v, got %v", expected, ranges)
	}
}

func TestAddSlots(t *testing.T) {
	slots := []Slot{1, 2, 3}
	addedSlots := []Slot{3, 4, 5}
	expected := []Slot{1, 2, 3, 4, 5}
	result := AddSlots(slots, addedSlots)
	sort.Sort(SlotSlice(result)) // Ensure the result is sorted before comparison
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestContains(t *testing.T) {
	slots := []Slot{1, 2, 3, 4, 5}
	if !Contains(slots, 3) {
		t.Errorf("expected to contain %v", 3)
	}
	if Contains(slots, 6) {
		t.Errorf("expected not to contain %v", 6)
	}
}

func TestBuildSlotSlice(t *testing.T) {
	min, max := Slot(1), Slot(5)
	expected := []Slot{1, 2, 3, 4, 5}
	result := BuildSlotSlice(min, max)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
