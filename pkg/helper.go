package pkg

func ToPointer[T any](v T) *T {
	return &v
}
