package domain

import (
	"fmt"
	"time"
)

type Order struct {
	id        int
	customer  string
	category  string
	amount    Money
	status    Status
	createdAt time.Time
}

func (o Order) ID() int {
	return o.id
}

func (o Order) Customer() string {
	return o.customer
}

func (o Order) Category() string {
	return o.category
}

func (o Order) Amount() Money {
	return o.amount
}

func (o Order) Status() Status {
	return o.status
}

func (o Order) CreatedAt() time.Time {
	return o.createdAt
}

func (o Order) IsPaid() bool {
	return o.status == StatusPaid
}

func NewOrder(id int, customer, category string, amount Money, status Status, createdAt time.Time) (Order, error) {
	order := Order{
		id:        id,
		customer:  customer,
		category:  category,
		amount:    amount,
		status:    status,
		createdAt: createdAt,
	}

	if err := order.validate(); err != nil {
		return Order{}, fmt.Errorf("domain/order: %w", err)
	}

	return order, nil
}

func (o Order) validate() error {
	if err := o.validateIDGreaterZero(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	if err := o.validateCustomerIsPresent(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	if err := o.validateCategoryIsPresent(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	if err := o.validateStatusIsPresent(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	if err := o.validateStatusIsValid(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	if err := o.validatePositiveAmount(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

func (o Order) validateIDGreaterZero() error {
	if o.id > 0 {
		return nil
	}

	return fmt.Errorf("validateIDGreaterZero: %w", ErrNonPositiveID)
}

func (o Order) validateCustomerIsPresent() error {
	if o.customer != "" {
		return nil
	}
	return fmt.Errorf("validateCustomerIsPresent: %w", ErrEmptyCustomer)
}

func (o Order) validateCategoryIsPresent() error {
	if o.category != "" {
		return nil
	}
	return fmt.Errorf("validateCategoryIsPresent: %w", ErrEmptyCategory)
}

func (o Order) validateStatusIsPresent() error {
	if o.status != "" {
		return nil
	}
	return fmt.Errorf("validateStatusIsPresent: %w", ErrEmptyStatus)
}

func (o Order) validateStatusIsValid() error {
	if o.status.Valid() {
		return nil
	}
	return fmt.Errorf("validateStatusIsValid: %w", ErrUnknownStatus)
}

func (o Order) validatePositiveAmount() error {
	if o.amount > 0 {
		return nil
	}
	return fmt.Errorf("validatePositiveAmount: %w", ErrNonPositiveAmount)
}
