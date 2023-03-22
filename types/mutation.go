package types

import "encoding/json"

type Mutation interface {
	json.Marshaler
	GetColumn() string
	GetOp() string
	GetValue() any
}

type mutationImpl[T BaseType] struct {
	column  string
	mutator string
	value   T
}

func (m mutationImpl[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{m.column, m.mutator, m.value})
}

func (m *mutationImpl[T]) GetColumn() string {
	return m.column
}

func (m *mutationImpl[T]) GetOp() string {
	return m.mutator
}

func (m *mutationImpl[T]) GetValue() any {
	return m.value
}

func Insert[T BaseType](column string, value T) Mutation {
	return &mutationImpl[T]{column, "insert", value}
}

func Delete[T BaseType](column string, value T) Mutation {
	return &mutationImpl[T]{column, "delete", value}
}

func Add[T BaseType](column string, value T) Mutation {
	return &mutationImpl[T]{column, "+=", value}
}

func Subtract[T BaseType](column string, value T) Mutation {
	return &mutationImpl[T]{column, "-=", value}
}

func Multiply[T BaseType](column string, value T) Mutation {
	return &mutationImpl[T]{column, "*=", value}
}

func Divide[T BaseType](column string, value T) Mutation {
	return &mutationImpl[T]{column, "/=", value}
}

func Modulo[T BaseType](column string, value T) Mutation {
	return &mutationImpl[T]{column, "%=", value}
}
