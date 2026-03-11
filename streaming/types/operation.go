package types

type Operation string

const (
	OperationInsert Operation = "INSERT"
	OperationUpdate Operation = "UPDATE"
	OperationDelete Operation = "DELETE"
)
