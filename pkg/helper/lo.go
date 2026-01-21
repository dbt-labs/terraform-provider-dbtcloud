package helper

// Additional functions to the samber/lo package.

// Same as Intersect but with a predicate function to compare elements.
func IntersectBy[T any](list1 []T, list2 []T, predicate func(T, T) bool) []T {
	result := []T{}

	for _, aValue := range list1 {
		found := false
		for _, bValue := range list2 {
			if predicate(aValue, bValue) {
				found = true
				break
			}
		}
		if found {
			result = append(result, aValue)
		}
	}

	return result
}

// Same as Difference but with a predicate function to compare elements.
func DifferenceBy[T any](
	list1 []T,
	list2 []T,
	predicate func(T, T) bool,
) ([]T, []T) {
	left := []T{}
	right := []T{}

	for _, aValue := range list1 {
		found := false
		for _, bValue := range list2 {
			if predicate(aValue, bValue) {
				found = true
				break
			}
		}
		if !found {
			left = append(left, aValue)
		}
	}

	for _, bValue := range list2 {
		found := false
		for _, aValue := range list1 {
			if predicate(bValue, aValue) {
				found = true
				break
			}
		}
		if !found {
			right = append(right, bValue)
		}
	}

	return left, right
}

// Same as Union but with a predicate function to compare elements.
// Note that items need to be unique already in list1 and list2
func UnionBy[T any](
	list1 []T,
	list2 []T,
	predicate func(T, T) bool,
) []T {
	response := []T{}

	response = append(response, list1...)

	for _, aValue := range list2 {
		found := false
		for _, bValue := range list1 {
			if predicate(aValue, bValue) {
				found = true
				break
			}
		}
		if !found {
			response = append(response, aValue)
		}
	}

	return response
}

// UniqBy removes duplicates from a slice based on a predicate function.
// This is useful for deduplicating slices where simple equality doesn't work.
func UniqBy[T any](
	list []T,
	predicate func(T, T) bool,
) []T {
	result := []T{}

	for _, item := range list {
		found := false
		for _, uniqueItem := range result {
			if predicate(item, uniqueItem) {
				found = true
				break
			}
		}
		if !found {
			result = append(result, item)
		}
	}

	return result
}
