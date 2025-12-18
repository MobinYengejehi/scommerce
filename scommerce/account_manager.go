package scommerce

import (
	"context"
	"errors"
	"time"

	"github.com/MobinYengejehi/scommerce/scommerce/otp"
)

var _ UserAccountManager[any] = &BuiltinUserAccountManager[any]{}

type userAccountManagerDatabase[AccountID comparable] interface {
	DBUserAccountManager[AccountID]
	userAccountDatabase[AccountID]
}

type BuiltinUserAccountManager[AccountID comparable] struct {
	DB                 userAccountManagerDatabase[AccountID]
	FS                 FileStorage
	OTP                *otp.OTP
	OTPTTL             time.Duration
	OrderStatusManager OrderStatusManager
}

func NewBuiltinUserAccountManager[AccountID comparable](
	db userAccountManagerDatabase[AccountID],
	fs FileStorage,
	codeLength int32,
	tokenLength int32,
	otpTTL time.Duration,
	osm OrderStatusManager,
) (*BuiltinUserAccountManager[AccountID], error) {
	otpDB, err := otp.NewInMemoryOTPDatabase()
	if err != nil {
		return nil, err
	}
	otpObj, err := otp.NewOTP(otpDB, codeLength, tokenLength)
	if err != nil {
		return nil, err
	}
	return &BuiltinUserAccountManager[AccountID]{
		DB:                 db,
		OTP:                otpObj,
		OTPTTL:             otpTTL,
		FS:                 fs,
		OrderStatusManager: osm,
	}, nil
}

func (accountManager *BuiltinUserAccountManager[AccountID]) newUserAccount(ctx context.Context, id AccountID, db userAccountDatabase[AccountID], form *UserAccountForm[AccountID]) (*BuiltinUserAccount[AccountID], error) {
	account := &BuiltinUserAccount[AccountID]{
		UserAccountForm: UserAccountForm[AccountID]{
			ID: id,
		},
		DB:                 db,
		FS:                 accountManager.FS,
		OrderStatusManager: accountManager.OrderStatusManager,
	}
	if err := account.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := account.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return account, nil
}

func (accountManager *BuiltinUserAccountManager[AccountID]) Authenticate(ctx context.Context, token string, password string, twoFactor string) (UserAccount[AccountID], error) {
	if res, err := accountManager.ValidateTwoFactor(ctx, token, twoFactor); err != nil {
		return nil, err
	} else if !res {
		return nil, errors.New("two factor code is not correct")
	}

	accountForm := UserAccountForm[AccountID]{}
	aid, err := accountManager.DB.AuthenticateUserAccount(ctx, token, password, &accountForm)
	if err != nil {
		return nil, err
	}
	account, err := accountManager.newUserAccount(ctx, aid, accountManager.DB, &accountForm)
	if err != nil {
		return nil, err
	}

	// this is getting checked in authenticate in db!
	// if res, err := account.IsBanned(ctx); err != nil {
	// 	return nil, err
	// } else if res != "" {
	// 	return nil, errors.New("account is banned for '" + res + "'")
	// }

	return account, nil
}

func (accountManager *BuiltinUserAccountManager[AccountID]) CancelTwoFactor(ctx context.Context, token string) error {
	return accountManager.OTP.Cancel(token)
}

func (accountManager *BuiltinUserAccountManager[AccountID]) Close(ctx context.Context) error {
	return accountManager.OTP.Clear()
}

func (accountManager *BuiltinUserAccountManager[AccountID]) GetAccount(ctx context.Context, token string) (UserAccount[AccountID], error) {
	accountForm := UserAccountForm[AccountID]{}
	aid, err := accountManager.DB.GetUserAccount(ctx, token, &accountForm)
	if err != nil {
		return nil, err
	}
	return accountManager.newUserAccount(ctx, aid, accountManager.DB, &accountForm)
}

func (accountManager *BuiltinUserAccountManager[AccountID]) GetAccountCount(ctx context.Context) (uint64, error) {
	return accountManager.DB.GetUserAccountCount(ctx)
}

func (accountManager *BuiltinUserAccountManager[AccountID]) GetAccountWithID(ctx context.Context, aid AccountID, fill bool) (UserAccount[AccountID], error) {
	if !fill {
		return accountManager.newUserAccount(ctx, aid, accountManager.DB, nil)
	}
	accountForm := UserAccountForm[AccountID]{}
	err := accountManager.DB.FillUserAccountWithID(ctx, aid, &accountForm)
	if err != nil {
		return nil, err
	}
	return accountManager.newUserAccount(ctx, aid, accountManager.DB, &accountForm)
}

func (accountManager *BuiltinUserAccountManager[AccountID]) GetAccounts(ctx context.Context, accounts []UserAccount[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserAccount[AccountID], error) {
	var err error = nil
	ids := make([]AccountID, 0, GetSafeLimit(limit))
	accountForms := make([]*UserAccountForm[AccountID], 0, cap(ids))
	ids, accountForms, err = accountManager.DB.GetUserAccounts(ctx, ids, accountForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	accs := accounts
	if accs == nil {
		accs = make([]UserAccount[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		account, err := accountManager.newUserAccount(ctx, ids[i], accountManager.DB, accountForms[i])
		if err != nil {
			return nil, err
		}
		accs = append(accs, account)
	}
	return accs, nil
}

func (accountManager *BuiltinUserAccountManager[AccountID]) Init(ctx context.Context) error {
	return accountManager.DB.InitUserAccountManager(ctx)
}

func (accountManager *BuiltinUserAccountManager[AccountID]) NewAccount(ctx context.Context, token string, password string, twoFactor string) (UserAccount[AccountID], error) {
	if res, err := accountManager.ValidateTwoFactor(ctx, token, twoFactor); err != nil {
		return nil, err
	} else if !res {
		return nil, errors.New("two factor code is not correct")
	}

	accountForm := UserAccountForm[AccountID]{}
	aid, err := accountManager.DB.NewUserAccount(ctx, token, password, &accountForm)
	if err != nil {
		return nil, err
	}

	return accountManager.newUserAccount(ctx, aid, accountManager.DB, &accountForm)
}

func (accountManager *BuiltinUserAccountManager[AccountID]) Pulse(ctx context.Context) error {
	return accountManager.OTP.Collect()
}

func (accountManager *BuiltinUserAccountManager[AccountID]) RemoveAccount(ctx context.Context, account UserAccount[AccountID]) error {
	aid, err := account.GetID(ctx)
	if err != nil {
		return err
	}
	return accountManager.DB.RemoveUserAccount(ctx, aid)
}

func (accountManager *BuiltinUserAccountManager[AccountID]) RemoveAccountWithToken(ctx context.Context, token string, password string, twoFactor string) error {
	if res, err := accountManager.ValidateTwoFactor(ctx, token, twoFactor); err != nil {
		return err
	} else if !res {
		return errors.New("two factor code is not correct")
	}
	return accountManager.DB.RemoveUserAccountWithToken(ctx, token, password)
}

func (accountManager *BuiltinUserAccountManager[AccountID]) RemoveAllAccounts(ctx context.Context) error {
	return accountManager.DB.RemoveAllUserAccounts(ctx)
}

func (accountManager *BuiltinUserAccountManager[AccountID]) RequestTwoFactor(ctx context.Context, token string) (string, error) {
	if exists, err := accountManager.OTP.Exists(token); err != nil {
		return "", err
	} else {
		if exists {
			return "", errors.New("otp session already exists")
		}
	}

	code, err := accountManager.OTP.NewCodeWithAssignedToken(token, accountManager.OTPTTL, 1)
	if err != nil {
		return "", err
	}

	return code, nil
}

func (accountManager *BuiltinUserAccountManager[AccountID]) ValidateTwoFactor(ctx context.Context, token string, code string) (bool, error) {
	result, err := accountManager.OTP.Validate(token, code)
	if err != nil {
		return false, errors.New("OTP Error: " + err.Error())
	}
	return result, nil
}

func (accountManager *BuiltinUserAccountManager[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinUserAccountManager[AccountID], error) {
	return accountManager, nil
}
