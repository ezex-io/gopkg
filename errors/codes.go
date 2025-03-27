package errors

var (
	ErrUnauthorized     = New(401, "unauthorized")
	ErrNotFound         = New(404, "not found")
	ErrPermissionDenied = New(403, "permission denied")
	ErrInvalidInput     = New(400, "invalid input")
	ErrInternal         = New(500, "internal error happened")
)
