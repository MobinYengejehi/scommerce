package otp

import (
	"errors"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"
)

const CodeKeys = "0123456789"
const TokenKeys = "abcdefghijklmnopqrstuvwxyz1234567890"

type OTP struct {
	DB          OTPDatabase
	CodeLength  atomic.Int32
	TokenLength atomic.Int32
	RandSrc     *rand.Rand
	MU          sync.Mutex
}

func NewOTP(db OTPDatabase, codeLength int32, tokenLength int32) (*OTP, error) {
	pcg := uint64(time.Now().UnixNano())
	rng := uint64(time.Now().UnixNano() >> 32)
	otp := &OTP{
		DB:      db,
		RandSrc: rand.New(rand.NewPCG(pcg, rng)),
	}
	otp.CodeLength.Store(codeLength)
	otp.TokenLength.Store(tokenLength)
	return otp, nil
}

func (otp *OTP) NewCode(ttl time.Duration, maxValidation int) (string, string, error) {
	const maxTry = 1e3

	tokenLength := otp.TokenLength.Load()
	codeLength := otp.CodeLength.Load()

	otp.MU.Lock()
	defer otp.MU.Unlock()

	token := string(GenerateRandomBytes(tokenLength, TokenKeys, otp.RandSrc))
	try := 0
	exists := false
	var err error = nil
	for exists, err = otp.DB.Has(token); exists && err == nil; exists, err = otp.DB.Has(token) {
		token = string(GenerateRandomBytes(tokenLength, TokenKeys, otp.RandSrc))
		try++
		if try > maxTry {
			return "", "", errors.New("exceed max tries")
		}
	}
	if err != nil {
		return "", "", err
	}

	code := string(GenerateRandomBytes(codeLength, CodeKeys, otp.RandSrc))
	value := OTPValue{
		Token:              token,
		Code:               code,
		CreatedAt:          time.Duration(time.Now().UnixNano()),
		TTL:                ttl,
		Validate:           0,
		MaxValidationCount: maxValidation,
	}
	if err := otp.DB.Save(token, value); err != nil {
		return "", "", err
	}

	return code, token, nil
}

func (otp *OTP) NewCodeWithAssignedToken(token string, ttl time.Duration, maxValidation int) (string, error) {
	otp.MU.Lock()
	defer otp.MU.Unlock()

	if exists, err := otp.DB.Has(token); err != nil {
		return "", err
	} else if exists {
		return "", errors.New("token already exists")
	}

	codeLength := otp.CodeLength.Load()

	code := string(GenerateRandomBytes(codeLength, CodeKeys, otp.RandSrc))
	value := OTPValue{
		Token:              token,
		Code:               code,
		CreatedAt:          time.Duration(time.Now().UnixNano()),
		TTL:                ttl,
		Validate:           0,
		MaxValidationCount: maxValidation,
	}
	if err := otp.DB.Save(token, value); err != nil {
		return "", err
	}

	return code, nil
}

func (otp *OTP) Refresh(token string) error {
	otp.MU.Lock()
	defer otp.MU.Unlock()

	value, err := otp.DB.Get(token)
	if err != nil {
		return err
	}

	value.Validate = 0
	value.CreatedAt = time.Duration(time.Now().UnixNano())

	return otp.DB.Save(token, value)
}

func (otp *OTP) Cancel(token string) error {
	otp.MU.Lock()
	defer otp.MU.Unlock()
	return otp.DB.Remove(token)
}

func (otp *OTP) Exists(token string) (bool, error) {
	otp.MU.Lock()
	defer otp.MU.Unlock()
	return otp.DB.Has(token)
}

func (otp *OTP) Validate(token string, code string) (bool, error) {
	otp.MU.Lock()
	value, err := otp.DB.Get(token)
	otp.MU.Unlock()
	if err != nil {
		return false, err
	}
	if value.Validate >= value.MaxValidationCount {
		otp.Cancel(token)
		return false, errors.New("max validation count exceeded")
	}
	if time.Duration(time.Now().UnixNano()) > value.CreatedAt+value.TTL {
		otp.Cancel(token)
		return false, errors.New("time to live exceeded")
	}

	if value.Code != code {
		return false, nil
	}

	otp.MU.Lock()
	defer otp.MU.Unlock()
	value.Validate++
	if err = otp.DB.Save(token, value); err != nil {
		return false, err
	}

	return true, nil
}

func (otp *OTP) Collect() error {
	count, err := otp.DB.ValueCount()
	if err != nil {
		return err
	}
	if count < 1 {
		return nil
	}

	values := make(map[string]OTPValue, count)
	if err := otp.DB.GetAll(values); err != nil {
		return err
	}

	for token, value := range values {
		if value.Validate >= value.MaxValidationCount {
			otp.Cancel(token)
			continue
		}
		if time.Duration(time.Now().UnixNano()) >= value.CreatedAt+value.TTL {
			otp.Cancel(token)
			continue
		}
	}

	return nil
}

func (otp *OTP) Clear() error {
	return otp.DB.Clear()
}

func GenerateRandomBytes(size int32, keys string, randSrc *rand.Rand) []byte {
	result := make([]byte, size)

	keysLen := len(keys)
	min := 0
	max := keysLen - 1

	for i := range size {
		randomIdx := randSrc.IntN(max-min+1) + min
		result[i] = keys[randomIdx]
	}

	return result
}
