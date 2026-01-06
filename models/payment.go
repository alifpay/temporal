package models

import "time"

const PaymentStatusSignal = "StatusSignal"

type Payment struct {
	ID       string
	Currency string
	Amount   float64
	PayDate  time.Time
}

type PaymentStatus struct {
	ID         string
	Status     string
	Code       int
	StatusDate time.Time
}
