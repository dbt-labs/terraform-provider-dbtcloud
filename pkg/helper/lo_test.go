package helper

import (
	"reflect"
	"testing"
)

func TestDifferenceBy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		list1         []int
		list2         []int
		predicate     func(list1 int, list2 int) bool
		expectedLeft  []int
		expectedRight []int
	}{
		{
			name:  "differences in left and right",
			list1: []int{0, 1, 2, 3, 4, 5},
			list2: []int{0, 2, 6},
			predicate: func(l int, r int) bool {
				return l == r
			},
			expectedLeft:  []int{1, 3, 4, 5},
			expectedRight: []int{6},
		},
		{
			name:  "no differences",
			list1: []int{0, 1, 2, 3, 4, 5},
			list2: []int{0, 1, 2, 3, 4, 5},
			predicate: func(l int, r int) bool {
				return l == r
			},
			expectedLeft:  []int{},
			expectedRight: []int{},
		},
		{
			name:  "differences in left only",
			list1: []int{0, 1, 2, 3, 4, 5},
			list2: []int{0, 1, 2},
			predicate: func(l int, r int) bool {
				return l == r
			},
			expectedLeft:  []int{3, 4, 5},
			expectedRight: []int{},
		},
		{
			name:  "differences in right only",
			list1: []int{0, 1, 2},
			list2: []int{0, 1, 2, 3, 4, 5},
			predicate: func(l int, r int) bool {
				return l == r
			},
			expectedLeft:  []int{},
			expectedRight: []int{3, 4, 5},
		},
		{
			name:  "differences in right only, (list1 is empty)",
			list1: []int{},
			list2: []int{0, 1, 2, 3, 4, 5},
			predicate: func(l int, r int) bool {
				return l == r
			},
			expectedLeft:  []int{},
			expectedRight: []int{0, 1, 2, 3, 4, 5},
		},
		{
			name:  "differences in left only (list2 is empty)",
			list1: []int{0, 1, 2, 3, 4, 5},
			list2: []int{},
			predicate: func(l int, r int) bool {
				return l == r
			},
			expectedLeft:  []int{0, 1, 2, 3, 4, 5},
			expectedRight: []int{},
		},
		{
			name:  "no differences (both list1 and list2 are empty)",
			list1: []int{},
			list2: []int{},
			predicate: func(l int, r int) bool {
				return l == r
			},
			expectedLeft:  []int{},
			expectedRight: []int{},
		},
		{
			name:  "no differences (both list1 and list2 are nil)",
			list1: nil,
			list2: nil,
			predicate: func(l int, r int) bool {
				return l == r
			},
			expectedLeft:  []int{},
			expectedRight: []int{},
		},
		{
			name:  "differences in right only (list1 is nil)",
			list1: nil,
			list2: []int{1, 2, 3},
			predicate: func(l int, r int) bool {
				return l == r
			},
			expectedLeft:  []int{},
			expectedRight: []int{1, 2, 3},
		},
		{
			name:  "differences in left only (list2 is nil)",
			list1: []int{1, 2, 3},
			list2: nil,
			predicate: func(l int, r int) bool {
				return l == r
			},
			expectedLeft:  []int{1, 2, 3},
			expectedRight: []int{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			left, right := DifferenceBy(
				testCase.list1,
				testCase.list2,
				func(l int, r int) bool {
					return l == r
				},
			)
			if !reflect.DeepEqual(left, testCase.expectedLeft) {
				t.Errorf(
					"Test case %s failed: expected left: %v, got: %v",
					testCase.name,
					testCase.expectedLeft,
					left,
				)
			}
			if !reflect.DeepEqual(right, testCase.expectedRight) {
				t.Errorf(
					"Test case %s failed: expected right: %v, got: %v",
					testCase.name,
					testCase.expectedRight,
					right,
				)
			}
		})
	}

}

func TestIntersectBy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		list1     []int
		list2     []int
		predicate func(list1 int, list2 int) bool
		expected  []int
	}{
		{
			name:  "differences in left and right",
			list1: []int{0, 1, 2, 3, 4, 5},
			list2: []int{0, 2, 6},
			predicate: func(l int, r int) bool {
				return l == r
			},
			expected: []int{0, 2},
		},
		{
			name:  "no differences",
			list1: []int{0, 1, 2, 3, 4, 5},
			list2: []int{0, 1, 2, 3, 4, 5},
			predicate: func(l int, r int) bool {
				return l == r
			},
			expected: []int{0, 1, 2, 3, 4, 5},
		},
		{
			name:  "list1 empty",
			list1: []int{},
			list2: []int{0, 1, 2, 3, 4, 5},
			predicate: func(l int, r int) bool {
				return l == r
			},
			expected: []int{},
		},
		{
			name:  "list2 empty",
			list1: []int{0, 1, 2, 3, 4, 5},
			list2: []int{},
			predicate: func(l int, r int) bool {
				return l == r
			},
			expected: []int{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := IntersectBy(
				testCase.list1,
				testCase.list2,
				func(l int, r int) bool {
					return l == r
				},
			)
			if !reflect.DeepEqual(result, testCase.expected) {
				t.Errorf(
					"Test case %s failed: expected left: %v, got: %v",
					testCase.name,
					testCase.expected,
					result,
				)
			}

		})
	}

}

func TestUnionBy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		list1     []int
		list2     []int
		predicate func(list1 int, list2 int) bool
		expected  []int
	}{
		{
			name:  "differences in left and right",
			list1: []int{0, 1, 2, 3, 4, 5},
			list2: []int{0, 2, 6},
			predicate: func(l int, r int) bool {
				return l == r
			},
			expected: []int{0, 1, 2, 3, 4, 5, 6},
		},
		{
			name:  "no differences",
			list1: []int{0, 1, 2, 3, 4, 5},
			list2: []int{0, 1, 2, 3, 4, 5},
			predicate: func(l int, r int) bool {
				return l == r
			},
			expected: []int{0, 1, 2, 3, 4, 5},
		},
		{
			name:  "list1 empty",
			list1: []int{},
			list2: []int{0, 1, 2, 3, 4, 5},
			predicate: func(l int, r int) bool {
				return l == r
			},
			expected: []int{0, 1, 2, 3, 4, 5},
		},
		{
			name:  "list2 empty",
			list1: []int{0, 1, 2, 3, 4, 5},
			list2: []int{},
			predicate: func(l int, r int) bool {
				return l == r
			},
			expected: []int{0, 1, 2, 3, 4, 5},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := UnionBy(
				testCase.list1,
				testCase.list2,
				func(l int, r int) bool {
					return l == r
				},
			)
			if !reflect.DeepEqual(result, testCase.expected) {
				t.Errorf(
					"Test case %s failed: expected left: %v, got: %v",
					testCase.name,
					testCase.expected,
					result,
				)
			}

		})
	}

}
