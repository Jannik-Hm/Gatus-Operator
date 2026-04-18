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

func FillIfNotValue[T comparable](existing T, fallback T, zero T) T {
	if existing == zero {
		return fallback
	}
	return existing
}
