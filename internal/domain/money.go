package domain

import "fmt"

type Money int64

func (m Money) String() string {
	wholeInt := m / 100
	fractionalInt := m % 100
	return fmt.Sprintf("%d.%02d", wholeInt, fractionalInt)
}
