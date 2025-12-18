package scommerce

import (
	"context"
	"sync"
)

var _ UserAddressManager[any] = &BuiltinUserAddressManager[any]{}
var _ UserAddress[any] = &BuiltinUserAddress[any]{}

type userAddressDatabase[AccountID comparable] interface {
	DBUserAddress[AccountID]
	DBCountry
}

type userAddressManagerDatabase[AccountID comparable] interface {
	DBUserAddressManager[AccountID]
	userAddressDatabase[AccountID]
}

type BuiltinUserAddressManager[AccountID comparable] struct {
	DB userAddressManagerDatabase[AccountID]
}

type UserAddressForm[AccountID comparable] struct {
	ID             uint64          `json:"id"`
	UserAccountID  AccountID       `json:"account_id"`
	AddressLine1   *string         `json:"address_line1,omitempty"`
	AddressLine2   *string         `json:"address_line2,omitempty"`
	City           *string         `json:"city,omitempty"`
	Country        *BuiltinCountry `json:"country,omitempty"`
	PostalCode     *string         `json:"postal_code,omitempty"`
	Region         *string         `json:"region,omitempty"`
	StreetNumber   *string         `json:"street_number,omitempty"`
	UnitNumber     *string         `json:"unit_number,omitempty"`
	IsDefaultState *bool           `json:"is_default,omitempty"`
}

type BuiltinUserAddress[AccountID comparable] struct {
	UserAddressForm[AccountID]
	DB userAddressDatabase[AccountID] `json:"-"`
	MU sync.RWMutex                   `json:"-"`
}

func NewBuiltinUserAddressManager[AccountID comparable](db userAddressManagerDatabase[AccountID]) *BuiltinUserAddressManager[AccountID] {
	return &BuiltinUserAddressManager[AccountID]{
		DB: db,
	}
}

func (addressManager *BuiltinUserAddressManager[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (addressManager *BuiltinUserAddressManager[AccountID]) GetAddressCount(ctx context.Context) (uint64, error) {
	return addressManager.DB.GetUserAddressCount(ctx)
}

func (addressManager *BuiltinUserAddressManager[AccountID]) GetAddressWithID(ctx context.Context, aid uint64, fill bool) (UserAddress[AccountID], error) {
	if !fill {
		var zeroAccountID AccountID
		return addressManager.newUserAddress(ctx, aid, zeroAccountID, addressManager.DB, nil)
	}
	addressForm := UserAddressForm[AccountID]{}
	err := addressManager.DB.FillUserAddressWithID(ctx, aid, &addressForm)
	if err != nil {
		return nil, err
	}
	return addressManager.newUserAddress(ctx, aid, addressForm.UserAccountID, addressManager.DB, &addressForm)
}

func (addressManager *BuiltinUserAddressManager[AccountID]) newUserAddress(ctx context.Context, id uint64, aid AccountID, db userAddressDatabase[AccountID], form *UserAddressForm[AccountID]) (*BuiltinUserAddress[AccountID], error) {
	addr := &BuiltinUserAddress[AccountID]{
		UserAddressForm: UserAddressForm[AccountID]{
			ID:            id,
			UserAccountID: aid,
		},
		DB: db,
	}
	if err := addr.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := addr.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return addr, nil
}

func (addressManager *BuiltinUserAddressManager[AccountID]) GetAddresses(ctx context.Context, addresses []UserAddress[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserAddress[AccountID], error) {
	var err error = nil
	ids := make([]DBUserAddressResult[AccountID], 0, GetSafeLimit(limit))
	addrForms := make([]*UserAddressForm[AccountID], 0, cap(ids))
	ids, addrForms, err = addressManager.DB.GetUserAddresses(ctx, ids, addrForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	adds := addresses
	if adds == nil {
		adds = make([]UserAddress[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		res := ids[i]
		addr, err := addressManager.newUserAddress(ctx, res.ID, res.AID, addressManager.DB, addrForms[i])
		if err != nil {
			return nil, err
		}
		adds = append(adds, addr)
	}
	return adds, nil
}

func (addressManager *BuiltinUserAddressManager[AccountID]) Init(ctx context.Context) error {
	return addressManager.DB.InitUserAddressManager(ctx)
}

func (addressManager *BuiltinUserAddressManager[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (addressManager *BuiltinUserAddressManager[AccountID]) RemoveAllAddresses(ctx context.Context) error {
	return addressManager.RemoveAllAddresses(ctx)
}

func (addressManager *BuiltinUserAddressManager[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinUserAddressManager[AccountID], error) {
	return addressManager, nil
}

func (address *BuiltinUserAddress[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (address *BuiltinUserAddress[AccountID]) GetAddressLine1(ctx context.Context) (string, error) {
	address.MU.RLock()
	if address.AddressLine1 != nil {
		defer address.MU.RUnlock()
		return *address.AddressLine1, nil
	}
	address.MU.RUnlock()
	id, err := address.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := address.UserAddressForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	addr, err := address.DB.GetUserAddressAddressLine1(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := address.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	address.MU.Lock()
	defer address.MU.Unlock()
	address.AddressLine1 = &addr
	return addr, nil
}

func (address *BuiltinUserAddress[AccountID]) GetAddressLine2(ctx context.Context) (string, error) {
	address.MU.RLock()
	if address.AddressLine2 != nil {
		defer address.MU.RUnlock()
		return *address.AddressLine2, nil
	}
	address.MU.RUnlock()
	id, err := address.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := address.UserAddressForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	addr, err := address.DB.GetUserAddressAddressLine2(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := address.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	address.MU.Lock()
	defer address.MU.Unlock()
	address.AddressLine2 = &addr
	return addr, nil
}

func (address *BuiltinUserAddress[AccountID]) GetCity(ctx context.Context) (string, error) {
	address.MU.RLock()
	if address.City != nil {
		defer address.MU.RUnlock()
		return *address.City, nil
	}
	address.MU.RUnlock()
	id, err := address.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := address.UserAddressForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	city, err := address.DB.GetUserAddressCity(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := address.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	address.MU.Lock()
	defer address.MU.Unlock()
	address.City = &city
	return city, nil
}

func (address *BuiltinUserAddress[AccountID]) GetCountry(ctx context.Context) (Country, error) {
	address.MU.RLock()
	if address.Country != nil {
		defer address.MU.RUnlock()
		return address.Country, nil
	}
	address.MU.RUnlock()
	id, err := address.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := address.UserAddressForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	countForm := CountryForm{}
	cid, err := address.DB.GetUserAddressCountry(ctx, &form, id, &countForm)
	if err != nil {
		return nil, err
	}
	if err := address.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	count := &BuiltinCountry{
		DB: address.DB,
		CountryForm: CountryForm{
			ID: cid,
		},
	}
	if err := count.Init(ctx); err != nil {
		return nil, err
	}
	if err := count.ApplyFormObject(ctx, &countForm); err != nil {
		return nil, err
	}
	address.MU.Lock()
	defer address.MU.Unlock()
	address.Country = count
	return count, nil
}

func (address *BuiltinUserAddress[AccountID]) GetID(ctx context.Context) (uint64, error) {
	address.MU.RLock()
	defer address.MU.RUnlock()
	return address.ID, nil
}

func (address *BuiltinUserAddress[AccountID]) GetPostalCode(ctx context.Context) (string, error) {
	address.MU.RLock()
	if address.PostalCode != nil {
		defer address.MU.RUnlock()
		return *address.PostalCode, nil
	}
	address.MU.RUnlock()
	id, err := address.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := address.UserAddressForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	code, err := address.DB.GetUserAddressPostalCode(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := address.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	address.MU.Lock()
	defer address.MU.Unlock()
	address.PostalCode = &code
	return code, nil
}

func (address *BuiltinUserAddress[AccountID]) GetRegion(ctx context.Context) (string, error) {
	address.MU.RLock()
	if address.Region != nil {
		defer address.MU.RUnlock()
		return *address.Region, nil
	}
	address.MU.RUnlock()
	id, err := address.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := address.UserAddressForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	region, err := address.DB.GetUserAddressRegion(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := address.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	address.MU.Lock()
	defer address.MU.Unlock()
	address.Region = &region
	return region, nil
}

func (address *BuiltinUserAddress[AccountID]) GetStreetNumber(ctx context.Context) (string, error) {
	address.MU.RLock()
	if address.StreetNumber != nil {
		defer address.MU.RUnlock()
		return *address.StreetNumber, nil
	}
	address.MU.RUnlock()
	id, err := address.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := address.UserAddressForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	number, err := address.DB.GetUserAddressStreetNumber(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := address.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	address.MU.Lock()
	defer address.MU.Unlock()
	address.StreetNumber = &number
	return number, nil
}

func (address *BuiltinUserAddress[AccountID]) GetUnitNumber(ctx context.Context) (string, error) {
	address.MU.RLock()
	if address.UnitNumber != nil {
		defer address.MU.RUnlock()
		return *address.UnitNumber, nil
	}
	address.MU.RUnlock()
	id, err := address.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := address.UserAddressForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	number, err := address.DB.GetUserAddressUnitNumber(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := address.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	address.MU.Lock()
	defer address.MU.Unlock()
	address.UnitNumber = &number
	return number, nil
}

func (address *BuiltinUserAddress[AccountID]) GetUserAccountID(ctx context.Context) (AccountID, error) {
	address.MU.RLock()
	defer address.MU.RUnlock()
	return address.UserAccountID, nil
}

func (address *BuiltinUserAddress[AccountID]) Init(ctx context.Context) error {
	return nil
}

func (address *BuiltinUserAddress[AccountID]) IsDefault(ctx context.Context) (bool, error) {
	address.MU.RLock()
	if address.IsDefaultState != nil {
		defer address.MU.RUnlock()
		return *address.IsDefaultState, nil
	}
	address.MU.RUnlock()
	id, err := address.GetID(ctx)
	if err != nil {
		return false, err
	}
	form, err := address.UserAddressForm.Clone(ctx)
	if err != nil {
		return false, err
	}
	isDefault, err := address.DB.IsUserAddressDefault(ctx, &form, id)
	if err != nil {
		return false, err
	}
	if err := address.ApplyFormObject(ctx, &form); err != nil {
		return false, err
	}
	address.MU.Lock()
	defer address.MU.Unlock()
	address.IsDefaultState = &isDefault
	return isDefault, nil
}

func (address *BuiltinUserAddress[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (address *BuiltinUserAddress[AccountID]) SetAddressLine1(ctx context.Context, addressLine string) error {
	id, err := address.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := address.UserAddressForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := address.DB.SetUserAddressAddressLine1(ctx, &form, id, addressLine); err != nil {
		return err
	}
	if err := address.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	address.MU.Lock()
	defer address.MU.Unlock()
	address.AddressLine1 = &addressLine
	return nil
}

func (address *BuiltinUserAddress[AccountID]) SetAddressLine2(ctx context.Context, addressLine string) error {
	id, err := address.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := address.UserAddressForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := address.DB.SetUserAddressAddressLine2(ctx, &form, id, addressLine); err != nil {
		return err
	}
	if err := address.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	address.MU.Lock()
	defer address.MU.Unlock()
	address.AddressLine2 = &addressLine
	return nil
}

func (address *BuiltinUserAddress[AccountID]) SetCity(ctx context.Context, city string) error {
	id, err := address.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := address.UserAddressForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := address.DB.SetUserAddressCity(ctx, &form, id, city); err != nil {
		return err
	}
	if err := address.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	address.MU.Lock()
	defer address.MU.Unlock()
	address.City = &city
	return nil
}

func (address *BuiltinUserAddress[AccountID]) SetCountry(ctx context.Context, country Country) error {
	var cid *uint64 = nil
	if country != nil {
		tcid, err := country.GetID(ctx)
		if err != nil {
			return err
		}
		cid = &tcid
	}
	id, err := address.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := address.UserAddressForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := address.DB.SetUserAddressCountry(ctx, &form, id, cid); err != nil {
		return err
	}
	if err := address.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	cont, err := country.ToBuiltinObject(ctx)
	if err != nil {
		return err
	}
	address.MU.Lock()
	defer address.MU.Unlock()
	address.Country = cont
	return nil
}

func (address *BuiltinUserAddress[AccountID]) SetDefault(ctx context.Context) error {
	id, err := address.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := address.UserAddressForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := address.DB.SetUserAddressDefault(ctx, &form, id); err != nil {
		return err
	}
	if err := address.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	address.MU.Lock()
	defer address.MU.Unlock()
	address.IsDefaultState = nil
	return nil
}

func (address *BuiltinUserAddress[AccountID]) SetPostalCode(ctx context.Context, code string) error {
	id, err := address.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := address.UserAddressForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := address.DB.SetUserAddressPostalCode(ctx, &form, id, code); err != nil {
		return err
	}
	if err := address.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	address.MU.Lock()
	defer address.MU.Unlock()
	address.PostalCode = &code
	return nil
}

func (address *BuiltinUserAddress[AccountID]) SetRegion(ctx context.Context, region string) error {
	id, err := address.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := address.UserAddressForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := address.DB.SetUserAddressRegion(ctx, &form, id, region); err != nil {
		return err
	}
	if err := address.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	address.MU.Lock()
	defer address.MU.Unlock()
	address.Region = &region
	return nil
}

func (address *BuiltinUserAddress[AccountID]) SetStreetNumber(ctx context.Context, streetNumber string) error {
	id, err := address.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := address.UserAddressForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := address.DB.SetUserAddressStreetNumber(ctx, &form, id, streetNumber); err != nil {
		return err
	}
	if err := address.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	address.MU.Lock()
	defer address.MU.Unlock()
	address.StreetNumber = &streetNumber
	return nil
}

func (address *BuiltinUserAddress[AccountID]) SetUnitNumber(ctx context.Context, unitNumber string) error {
	id, err := address.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := address.UserAddressForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := address.DB.SetUserAddressUnitNumber(ctx, &form, id, unitNumber); err != nil {
		return err
	}
	if err := address.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	address.MU.Lock()
	defer address.MU.Unlock()
	address.UnitNumber = &unitNumber
	return nil
}

func (address *BuiltinUserAddress[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinUserAddress[AccountID], error) {
	return address, nil
}

func (address *BuiltinUserAddress[AccountID]) ToFormObject(ctx context.Context) (*UserAddressForm[AccountID], error) {
	address.MU.RLock()
	defer address.MU.RUnlock()
	return &address.UserAddressForm, nil
}

func (address *BuiltinUserAddress[AccountID]) ApplyFormObject(ctx context.Context, form *UserAddressForm[AccountID]) error {
	address.MU.Lock()
	defer address.MU.Unlock()
	// Conditional copy: only update non-zero IDs and non-nil pointers
	if form.ID != 0 {
		address.ID = form.ID
	}
	// Check if UserAccountID is zero value (requires generic type comparison)
	var zeroAccountID AccountID
	if form.UserAccountID != zeroAccountID {
		address.UserAccountID = form.UserAccountID
	}
	if form.AddressLine1 != nil {
		address.AddressLine1 = form.AddressLine1
	}
	if form.AddressLine2 != nil {
		address.AddressLine2 = form.AddressLine2
	}
	if form.City != nil {
		address.City = form.City
	}
	if form.Country != nil {
		address.Country = form.Country
	}
	if form.PostalCode != nil {
		address.PostalCode = form.PostalCode
	}
	if form.Region != nil {
		address.Region = form.Region
	}
	if form.StreetNumber != nil {
		address.StreetNumber = form.StreetNumber
	}
	if form.UnitNumber != nil {
		address.UnitNumber = form.UnitNumber
	}
	if form.IsDefaultState != nil {
		address.IsDefaultState = form.IsDefaultState
	}
	return nil
}

func (form *UserAddressForm[AccountID]) Clone(ctx context.Context) (UserAddressForm[AccountID], error) {
	var cloned UserAddressForm[AccountID] = *form
	return cloned, nil
}
