package retry

type (
	Task         func() error
	TaskT[T any] func() (T, error)
)
