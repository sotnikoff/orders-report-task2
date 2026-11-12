package domain

import "errors"

var (
	ErrNonPositiveID     = errors.New("ID должен быть больше 0")
	ErrEmptyCustomer     = errors.New("клиент - пустое значение")
	ErrEmptyCategory     = errors.New("категория - пустое значение")
	ErrEmptyStatus       = errors.New("статус - пустое значение")
	ErrUnknownStatus     = errors.New("неизвестный статус")
	ErrNonPositiveAmount = errors.New("сумма должна быть строго больше нуля")
	ErrDuplicateID       = errors.New("ID повторяющийся")
)
