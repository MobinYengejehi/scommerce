package scommerce

import (
	"context"
	"sync"
)

type countryDatabase interface {
	DBCountryManager
	DBCountry
}

var _ CountryManager = &BuiltinCountryManager{}
var _ Country = &BuiltinCountry{}

type BuiltinCountryManager struct {
	DB countryDatabase
}

type CountryForm struct {
	ID   uint64  `json:"id"`
	Name *string `json:"name,omitempty"`
}

type BuiltinCountry struct {
	CountryForm
	DB DBCountry    `json:"-"`
	MU sync.RWMutex `json:"-"`
}

func NewBuiltinCountryManager(db countryDatabase) *BuiltinCountryManager {
	return &BuiltinCountryManager{
		DB: db,
	}
}

func (countryManager *BuiltinCountryManager) newBuiltinCountry(ctx context.Context, id uint64, db DBCountry, form *CountryForm) (*BuiltinCountry, error) {
	country := &BuiltinCountry{
		CountryForm: CountryForm{
			ID: id,
		},
		DB: db,
	}
	if err := country.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := country.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return country, nil
}

func (countryManager *BuiltinCountryManager) Close(ctx context.Context) error {
	return nil
}

func (countryManager *BuiltinCountryManager) ExistsCountry(ctx context.Context, name string) (bool, error) {
	return countryManager.DB.ExistsCountry(ctx, name)
}

func (countryManager *BuiltinCountryManager) GetCountryByName(ctx context.Context, name string) (Country, error) {
	countForm := CountryForm{}
	id, err := countryManager.DB.GetCountryByName(ctx, name, &countForm)
	if err != nil {
		return nil, err
	}
	return countryManager.newBuiltinCountry(ctx, id, countryManager.DB, &countForm)
}

func (countryManager *BuiltinCountryManager) GetCountryCount(ctx context.Context) (uint64, error) {
	return countryManager.DB.GetCountryCount(ctx)
}

func (countryManager *BuiltinCountryManager) GetCountries(ctx context.Context, countries []Country, skip int64, limit int64, queueOrder QueueOrder) ([]Country, error) {
	var err error = nil
	ids := make([]uint64, 0, GetSafeLimit(limit))
	countForms := make([]*CountryForm, 0, cap(ids))
	ids, countForms, err = countryManager.DB.GetCountries(ctx, ids, countForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	cons := countries
	if cons == nil {
		cons = make([]Country, 0, len(ids))
	}
	for i := range len(ids) {
		con, err := countryManager.newBuiltinCountry(ctx, ids[i], countryManager.DB, countForms[i])
		if err != nil {
			return nil, err
		}
		cons = append(cons, con)
	}
	return cons, nil
}

func (countryManager *BuiltinCountryManager) Init(ctx context.Context) error {
	return countryManager.DB.InitCountryManager(ctx)
}

func (countryManager *BuiltinCountryManager) NewCountry(ctx context.Context, name string) (Country, error) {
	countForm := CountryForm{}
	id, err := countryManager.DB.NewCountry(ctx, name, &countForm)
	if err != nil {
		return nil, err
	}
	return countryManager.newBuiltinCountry(ctx, id, countryManager.DB, &countForm)
}

func (countryManager *BuiltinCountryManager) Pulse(ctx context.Context) error {
	return nil
}

func (countryManager *BuiltinCountryManager) RemoveAllCountries(ctx context.Context) error {
	return countryManager.DB.RemoveAllCountries(ctx)
}

func (countryManager *BuiltinCountryManager) RemoveCountry(ctx context.Context, country Country) error {
	id, err := country.GetID(ctx)
	if err != nil {
		return err
	}
	return countryManager.DB.RemoveCountry(ctx, id)
}

func (countryManager *BuiltinCountryManager) ToBuiltinObject(ctx context.Context) (*BuiltinCountryManager, error) {
	return countryManager, nil
}

func (country *BuiltinCountry) Close(ctx context.Context) error {
	return nil
}

func (country *BuiltinCountry) GetID(ctx context.Context) (uint64, error) {
	country.MU.RLock()
	defer country.MU.RUnlock()
	return country.ID, nil
}

func (country *BuiltinCountry) GetName(ctx context.Context) (string, error) {
	country.MU.RLock()
	if country.Name != nil {
		defer country.MU.RUnlock()
		return *country.Name, nil
	}
	country.MU.RUnlock()
	id, err := country.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := country.CountryForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	name, err := country.DB.GetCountryName(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := country.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	country.MU.Lock()
	defer country.MU.Unlock()
	country.Name = &name
	return name, nil
}

func (country *BuiltinCountry) Init(ctx context.Context) error {
	return nil
}

func (country *BuiltinCountry) Pulse(ctx context.Context) error {
	return nil
}

func (country *BuiltinCountry) SetName(ctx context.Context, name string) error {
	id, err := country.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := country.CountryForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := country.DB.SetCountryName(ctx, &form, id, name); err != nil {
		return err
	}
	if err := country.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	country.MU.Lock()
	defer country.MU.Unlock()
	country.Name = &name
	return nil
}

func (country *BuiltinCountry) ToBuiltinObject(ctx context.Context) (*BuiltinCountry, error) {
	return country, nil
}

func (country *BuiltinCountry) ToFormObject(ctx context.Context) (*CountryForm, error) {
	country.MU.RLock()
	defer country.MU.RUnlock()
	return &country.CountryForm, nil
}

func (country *BuiltinCountry) ApplyFormObject(ctx context.Context, form *CountryForm) error {
	country.MU.Lock()
	defer country.MU.Unlock()
	// Conditional copy: only update non-zero IDs and non-nil pointers
	if form.ID != 0 {
		country.ID = form.ID
	}
	if form.Name != nil {
		country.Name = form.Name
	}
	return nil
}

func (form *CountryForm) Clone(ctx context.Context) (CountryForm, error) {
	var cloned CountryForm = *form
	return cloned, nil
}
