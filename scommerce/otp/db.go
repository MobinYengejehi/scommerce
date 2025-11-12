package otp

import (
	"errors"
	"time"
)

var _ OTPDatabase = &InMemoryOTPDatabase{}

type OTPValue struct {
	Token              string
	Code               string
	CreatedAt          time.Duration
	TTL                time.Duration
	Validate           int
	MaxValidationCount int
}

type OTPDatabase interface {
	Save(token string, value OTPValue) error
	Get(token string) (OTPValue, error)
	GetAll(values map[string]OTPValue) error
	ValueCount() (int, error)
	Has(token string) (bool, error)
	Remove(token string) error
	Clear() error
}

// OTP object has already guaranteed the race conditions. so we don't need to use atomics for our db.
type InMemoryOTPDatabase struct {
	values map[string]OTPValue
}

func NewInMemoryOTPDatabase() (*InMemoryOTPDatabase, error) {
	return &InMemoryOTPDatabase{
		values: make(map[string]OTPValue),
	}, nil
}

func (db *InMemoryOTPDatabase) Get(token string) (OTPValue, error) {
	if v, exists := db.values[token]; exists {
		return v, nil
	}
	return OTPValue{}, errors.New("doesn't exist")
}

func (db *InMemoryOTPDatabase) GetAll(values map[string]OTPValue) error {
	for token, value := range db.values {
		values[token] = value
	}
	return nil
}

func (db *InMemoryOTPDatabase) ValueCount() (int, error) {
	return len(db.values), nil
}

func (db *InMemoryOTPDatabase) Has(token string) (bool, error) {
	_, exists := db.values[token]
	return exists, nil
}

func (db *InMemoryOTPDatabase) Save(token string, code OTPValue) error {
	db.values[token] = code
	return nil
}

func (db *InMemoryOTPDatabase) Remove(token string) error {
	delete(db.values, token)
	return nil
}

func (db *InMemoryOTPDatabase) Clear() error {
	db.values = make(map[string]OTPValue)
	return nil
}
