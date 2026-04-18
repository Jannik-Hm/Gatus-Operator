package gatusconfig

func clonePtr[T any](v *T) *T {
	if v == nil {
		return nil
	}
	c := *v
	return &c
}

type Cloner[T any] interface {
	Clone() T
}

func listClone[Source Cloner[Source]](list []Source) []Source {
	if list == nil {
		return nil
	}

	result := make([]Source, len(list))
	for i, a := range list {
		result[i] = a.Clone()
	}
	return result
}

type Merger[T any] interface {
	Merge(...T) T
}

func FillIfValue[T comparable](existing T, fallback T, value T) T {
	if existing == value {
		return fallback
	}
	return existing
}

func MergeIntoList[T any](existing []T, fallback []T) {
	for _, element := range fallback {
		existing = append(existing, element)
	}
}

func MergeIntoListUnique[T comparable](existing []T, fallback []T) {
	existing_values := map[T]struct{}{}
	for _, value := range existing {
		existing_values[value] = struct{}{}
	}
	for _, value := range fallback {
		if _, ok := existing_values[value]; !ok {
			existing = append(existing, value)
		}
	}
}

func MergeIntoMap[Key comparable, Value any](existing map[Key]Value, fallback map[Key]Value) {
	for key, value := range fallback {
		if _, ok := existing[key]; !ok {
			existing[key] = value
		}
	}
}
