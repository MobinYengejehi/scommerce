package scommerce

import (
	"context"
	"sync"
)

type userRoleDatabase interface {
	DBUserRoleManager
	DBUserRole
}

var _ UserRoleManager = &BuiltinUserRoleManager{}
var _ UserRole = &BuiltinUserRole{}

type BuiltinUserRoleManager struct {
	DB userRoleDatabase
}

type UserRoleForm struct {
	ID   uint64  `json:"id"`
	Name *string `json:"name,omitempty"`
}

type BuiltinUserRole struct {
	UserRoleForm
	DB DBUserRole   `json:"-"`
	MU sync.RWMutex `json:"-"`
}

func NewBuiltinUserRoleManager(db userRoleDatabase) *BuiltinUserRoleManager {
	return &BuiltinUserRoleManager{
		DB: db,
	}
}

func (userRoleManager *BuiltinUserRoleManager) newBuiltinUserRole(ctx context.Context, id uint64, db DBUserRole, form *UserRoleForm) (*BuiltinUserRole, error) {
	role := &BuiltinUserRole{
		UserRoleForm: UserRoleForm{
			ID: id,
		},
		DB: db,
	}
	if err := role.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := role.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return role, nil
}

func (userRoleManager *BuiltinUserRoleManager) Close(ctx context.Context) error {
	return nil
}

func (userRoleManager *BuiltinUserRoleManager) ExistsUserRole(ctx context.Context, name string) (bool, error) {
	return userRoleManager.DB.ExistsUserRole(ctx, name)
}

func (userRoleManager *BuiltinUserRoleManager) GetUserRoleByName(ctx context.Context, name string) (UserRole, error) {
	roleForm := UserRoleForm{}
	id, err := userRoleManager.DB.GetUserRoleByName(ctx, name, &roleForm)
	if err != nil {
		return nil, err
	}
	return userRoleManager.newBuiltinUserRole(ctx, id, userRoleManager.DB, &roleForm)
}

func (userRoleManager *BuiltinUserRoleManager) GetUserRoleCount(ctx context.Context) (uint64, error) {
	return userRoleManager.DB.GetUserRoleCount(ctx)
}

func (userRoleManager *BuiltinUserRoleManager) GetUserRoles(ctx context.Context, userRoles []UserRole, skip int64, limit int64, queueOrder QueueOrder) ([]UserRole, error) {
	var err error = nil
	ids := make([]uint64, 0, GetSafeLimit(limit))
	roleForms := make([]*UserRoleForm, 0, cap(ids))
	ids, roleForms, err = userRoleManager.DB.GetUserRoles(ctx, ids, roleForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	types := userRoles
	if types == nil {
		types = make([]UserRole, 0, len(ids))
	}
	for i := range len(ids) {
		pType, err := userRoleManager.newBuiltinUserRole(ctx, ids[i], userRoleManager.DB, roleForms[i])
		if err != nil {
			return nil, err
		}
		types = append(types, pType)
	}
	return types, nil
}

func (userRoleManager *BuiltinUserRoleManager) GetRoleWithID(ctx context.Context, rid uint64, fill bool) (UserRole, error) {
	if !fill {
		return userRoleManager.newBuiltinUserRole(ctx, rid, userRoleManager.DB, nil)
	}
	roleForm := UserRoleForm{}
	err := userRoleManager.DB.FillUserRoleWithID(ctx, rid, &roleForm)
	if err != nil {
		return nil, err
	}
	return userRoleManager.newBuiltinUserRole(ctx, rid, userRoleManager.DB, &roleForm)
}

func (userRoleManager *BuiltinUserRoleManager) Init(ctx context.Context) error {
	return userRoleManager.DB.InitUserRoleManager(ctx)
}

func (userRoleManager *BuiltinUserRoleManager) NewUserRole(ctx context.Context, name string) (UserRole, error) {
	roleForm := UserRoleForm{}
	id, err := userRoleManager.DB.NewUserRole(ctx, name, &roleForm)
	if err != nil {
		return nil, err
	}
	return userRoleManager.newBuiltinUserRole(ctx, id, userRoleManager.DB, &roleForm)
}

func (userRoleManager *BuiltinUserRoleManager) Pulse(ctx context.Context) error {
	return nil
}

func (userRoleManager *BuiltinUserRoleManager) RemoveAllUserRoles(ctx context.Context) error {
	return userRoleManager.DB.RemoveAllUserRoles(ctx)
}

func (userRoleManager *BuiltinUserRoleManager) RemoveUserRole(ctx context.Context, userRole UserRole) error {
	id, err := userRole.GetID(ctx)
	if err != nil {
		return err
	}
	return userRoleManager.DB.RemoveUserRole(ctx, id)
}

func (userRoleManager *BuiltinUserRoleManager) ToBuiltinObject(ctx context.Context) (*BuiltinUserRoleManager, error) {
	return userRoleManager, nil
}

func (userRole *BuiltinUserRole) Close(ctx context.Context) error {
	return nil
}

func (userRole *BuiltinUserRole) GetID(ctx context.Context) (uint64, error) {
	userRole.MU.RLock()
	defer userRole.MU.RUnlock()
	return userRole.ID, nil
}

func (userRole *BuiltinUserRole) GetName(ctx context.Context) (string, error) {
	userRole.MU.RLock()
	if userRole.Name != nil {
		defer userRole.MU.RUnlock()
		return *userRole.Name, nil
	}
	userRole.MU.RUnlock()
	id, err := userRole.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := userRole.UserRoleForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	name, err := userRole.DB.GetUserRoleName(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := userRole.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	userRole.MU.Lock()
	defer userRole.MU.Unlock()
	userRole.Name = &name
	return name, nil
}

func (userRole *BuiltinUserRole) Init(ctx context.Context) error {
	return nil
}

func (userRole *BuiltinUserRole) Pulse(ctx context.Context) error {
	return nil
}

func (userRole *BuiltinUserRole) SetName(ctx context.Context, name string) error {
	id, err := userRole.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := userRole.UserRoleForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := userRole.DB.SetUserRoleName(ctx, &form, id, name); err != nil {
		return err
	}
	if err := userRole.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	userRole.MU.Lock()
	defer userRole.MU.Unlock()
	userRole.Name = &name
	return nil
}

func (userRole *BuiltinUserRole) ToBuiltinObject(ctx context.Context) (*BuiltinUserRole, error) {
	return userRole, nil
}

func (userRole *BuiltinUserRole) ToFormObject(ctx context.Context) (*UserRoleForm, error) {
	userRole.MU.RLock()
	defer userRole.MU.RUnlock()
	return &userRole.UserRoleForm, nil
}

func (userRole *BuiltinUserRole) ApplyFormObject(ctx context.Context, form *UserRoleForm) error {
	userRole.MU.Lock()
	defer userRole.MU.Unlock()
	// Conditional copy: only update non-zero IDs and non-nil pointers
	if form.ID != 0 {
		userRole.ID = form.ID
	}
	if form.Name != nil {
		userRole.Name = form.Name
	}
	return nil
}

func (form *UserRoleForm) Clone(ctx context.Context) (UserRoleForm, error) {
	var cloned UserRoleForm = *form
	return cloned, nil
}
