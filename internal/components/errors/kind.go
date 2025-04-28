package errors

// Kind is an error representation of a mapped error.
type Kind string

var (
	KindValidation   Kind = "ValidationError"
	KindInternal     Kind = "InternalError"
	KindNotFound     Kind = "NotFoundError"
	KindPrecondition Kind = "ConditionError"
	KindPermission   Kind = "PermissionError"
	KindRPC          Kind = "RPCError"
	KindCustom       Kind = "CustomError"
)

func (k Kind) String() string {
	return string(k)
}
