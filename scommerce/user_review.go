package scommerce

import (
	"context"
	"sync"
)

var _ UserReviewManager[any] = &BuiltinUserReviewManager[any]{}
var _ UserReview[any] = &BuiltinUserReview[any]{}

type userReviewManagerDatabase[AccountID comparable] interface {
	DBUserReviewManager[AccountID]
	DBUserReview[AccountID]
	productItemDatabase[AccountID]
}

type userReviewDatabase[AccountID comparable] interface {
	DBUserReview[AccountID]
	productItemDatabase[AccountID]
}

type BuiltinUserReviewManager[AccountID comparable] struct {
	DB userReviewManagerDatabase[AccountID]
	FS FileStorage
}

type UserReviewForm[AccountID comparable] struct {
	ID            uint64                         `json:"id"`
	UserAccountID AccountID                      `json:"user_account_id"`
	RatingValue   *int32                         `json:"rating_value,omitempty"`
	Comment       *string                        `json:"comment,omitempty"`
	ProductItem   *BuiltinProductItem[AccountID] `json:"product_item,omitempty"`
}

type BuiltinUserReview[AccountID comparable] struct {
	UserReviewForm[AccountID]
	DB userReviewDatabase[AccountID] `json:"-"`
	FS FileStorage                   `json:"-"`
	MU sync.RWMutex                  `json:"-"`
}

func NewBuiltinUserReviewManager[AccountID comparable](db userReviewManagerDatabase[AccountID], fs FileStorage) *BuiltinUserReviewManager[AccountID] {
	return &BuiltinUserReviewManager[AccountID]{
		DB: db,
		FS: fs,
	}
}

func (reviewManager *BuiltinUserReviewManager[AccountID]) newUserReview(ctx context.Context, id uint64, db userReviewDatabase[AccountID], form *UserReviewForm[AccountID]) (*BuiltinUserReview[AccountID], error) {
	var zeroAccountID AccountID
	review := &BuiltinUserReview[AccountID]{
		UserReviewForm: UserReviewForm[AccountID]{
			ID:            id,
			UserAccountID: zeroAccountID,
		},
		DB: db,
		FS: reviewManager.FS,
	}
	if err := review.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := review.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return review, nil
}

func (reviewManager *BuiltinUserReviewManager[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (reviewManager *BuiltinUserReviewManager[AccountID]) GetUserReviewCount(ctx context.Context) (uint64, error) {
	return reviewManager.DB.GetUserReviewCount(ctx)
}

func (reviewManager *BuiltinUserReviewManager[AccountID]) GetUserReviewWithID(ctx context.Context, rid uint64, fill bool) (UserReview[AccountID], error) {
	if !fill {
		return reviewManager.newUserReview(ctx, rid, reviewManager.DB, nil)
	}
	reviewForm := UserReviewForm[AccountID]{}
	err := reviewManager.DB.FillUserReviewWithID(ctx, rid, &reviewForm)
	if err != nil {
		return nil, err
	}
	return reviewManager.newUserReview(ctx, rid, reviewManager.DB, &reviewForm)
}

func (reviewManager *BuiltinUserReviewManager[AccountID]) GetUserReviews(ctx context.Context, reviews []UserReview[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserReview[AccountID], error) {
	var err error = nil
	ids := make([]DBUserReviewResult[AccountID], 0, GetSafeLimit(limit))
	reviewForms := make([]*UserReviewForm[AccountID], 0, cap(ids))
	ids, reviewForms, err = reviewManager.DB.GetUserReviews(ctx, ids, reviewForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	revs := reviews
	if revs == nil {
		revs = make([]UserReview[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		review, err := reviewManager.newUserReview(ctx, ids[i].ID, reviewManager.DB, reviewForms[i])
		if err != nil {
			return nil, err
		}
		revs = append(revs, review)
	}
	return revs, nil
}

func (reviewManager *BuiltinUserReviewManager[AccountID]) GetUserReviewsForAccount(ctx context.Context, account UserAccount[AccountID], reviews []UserReview[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserReview[AccountID], error) {
	var err error = nil
	aid, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]DBUserReviewResult[AccountID], 0, GetSafeLimit(limit))
	reviewForms := make([]*UserReviewForm[AccountID], 0, cap(ids))
	ids, reviewForms, err = reviewManager.DB.GetUserReviewsForAccount(ctx, aid, ids, reviewForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	revs := reviews
	if revs == nil {
		revs = make([]UserReview[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		review, err := reviewManager.newUserReview(ctx, ids[i].ID, reviewManager.DB, reviewForms[i])
		if err != nil {
			return nil, err
		}
		revs = append(revs, review)
	}
	return revs, nil
}

func (reviewManager *BuiltinUserReviewManager[AccountID]) GetUserReviewsForProductItem(ctx context.Context, productItem ProductItem[AccountID], reviews []UserReview[AccountID], skip int64, limit int64, queueOrder QueueOrder) ([]UserReview[AccountID], error) {
	var err error = nil
	pid, err := productItem.GetID(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]DBUserReviewResult[AccountID], 0, GetSafeLimit(limit))
	reviewForms := make([]*UserReviewForm[AccountID], 0, cap(ids))
	ids, reviewForms, err = reviewManager.DB.GetUserReviewsForProductItem(ctx, pid, ids, reviewForms, skip, limit, queueOrder)
	if err != nil {
		return nil, err
	}
	revs := reviews
	if revs == nil {
		revs = make([]UserReview[AccountID], 0, len(ids))
	}
	for i := range len(ids) {
		review, err := reviewManager.newUserReview(ctx, ids[i].ID, reviewManager.DB, reviewForms[i])
		if err != nil {
			return nil, err
		}
		revs = append(revs, review)
	}
	return revs, nil
}

func (reviewManager *BuiltinUserReviewManager[AccountID]) Init(ctx context.Context) error {
	return reviewManager.DB.InitUserReviewManager(ctx)
}

func (reviewManager *BuiltinUserReviewManager[AccountID]) NewUserReview(ctx context.Context, account UserAccount[AccountID], productItem ProductItem[AccountID], ratingValue int32, comment string) (UserReview[AccountID], error) {
	aid, err := account.GetID(ctx)
	if err != nil {
		return nil, err
	}
	pid, err := productItem.GetID(ctx)
	if err != nil {
		return nil, err
	}
	reviewForm := UserReviewForm[AccountID]{}
	rid, err := reviewManager.DB.NewUserReview(ctx, aid, pid, ratingValue, comment, &reviewForm)
	if err != nil {
		return nil, err
	}
	return reviewManager.newUserReview(ctx, rid, reviewManager.DB, &reviewForm)
}

func (reviewManager *BuiltinUserReviewManager[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (reviewManager *BuiltinUserReviewManager[AccountID]) RemoveAllUserReviews(ctx context.Context) error {
	return reviewManager.DB.RemoveAllUserReviews(ctx)
}

func (reviewManager *BuiltinUserReviewManager[AccountID]) RemoveUserReview(ctx context.Context, review UserReview[AccountID]) error {
	rid, err := review.GetID(ctx)
	if err != nil {
		return err
	}
	return reviewManager.DB.RemoveUserReview(ctx, rid)
}

func (reviewManager *BuiltinUserReviewManager[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinUserReviewManager[AccountID], error) {
	return reviewManager, nil
}

func (review *BuiltinUserReview[AccountID]) Close(ctx context.Context) error {
	return nil
}

func (review *BuiltinUserReview[AccountID]) GetID(ctx context.Context) (uint64, error) {
	review.MU.RLock()
	defer review.MU.RUnlock()
	return review.ID, nil
}

func (review *BuiltinUserReview[AccountID]) GetUserAccountID(ctx context.Context) (AccountID, error) {
	review.MU.RLock()
	defer review.MU.RUnlock()
	return review.UserAccountID, nil
}

func (review *BuiltinUserReview[AccountID]) GetRatingValue(ctx context.Context) (int32, error) {
	review.MU.RLock()
	if review.RatingValue != nil {
		defer review.MU.RUnlock()
		return *review.RatingValue, nil
	}
	review.MU.RUnlock()
	id, err := review.GetID(ctx)
	if err != nil {
		return 0, err
	}
	form, err := review.UserReviewForm.Clone(ctx)
	if err != nil {
		return 0, err
	}
	rating, err := review.DB.GetUserReviewRatingValue(ctx, &form, id)
	if err != nil {
		return 0, err
	}
	if err := review.ApplyFormObject(ctx, &form); err != nil {
		return 0, err
	}
	review.MU.Lock()
	defer review.MU.Unlock()
	review.RatingValue = &rating
	return rating, nil
}

func (review *BuiltinUserReview[AccountID]) SetRatingValue(ctx context.Context, rating int32) error {
	id, err := review.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := review.UserReviewForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := review.DB.SetUserReviewRatingValue(ctx, &form, id, rating); err != nil {
		return err
	}
	if err := review.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	review.MU.Lock()
	defer review.MU.Unlock()
	review.RatingValue = &rating
	return nil
}

func (review *BuiltinUserReview[AccountID]) GetComment(ctx context.Context) (string, error) {
	review.MU.RLock()
	if review.Comment != nil {
		defer review.MU.RUnlock()
		return *review.Comment, nil
	}
	review.MU.RUnlock()
	id, err := review.GetID(ctx)
	if err != nil {
		return "", err
	}
	form, err := review.UserReviewForm.Clone(ctx)
	if err != nil {
		return "", err
	}
	comment, err := review.DB.GetUserReviewComment(ctx, &form, id)
	if err != nil {
		return "", err
	}
	if err := review.ApplyFormObject(ctx, &form); err != nil {
		return "", err
	}
	review.MU.Lock()
	defer review.MU.Unlock()
	review.Comment = &comment
	return comment, nil
}

func (review *BuiltinUserReview[AccountID]) SetComment(ctx context.Context, comment string) error {
	id, err := review.GetID(ctx)
	if err != nil {
		return err
	}
	form, err := review.UserReviewForm.Clone(ctx)
	if err != nil {
		return err
	}
	if err := review.DB.SetUserReviewComment(ctx, &form, id, comment); err != nil {
		return err
	}
	if err := review.ApplyFormObject(ctx, &form); err != nil {
		return err
	}
	review.MU.Lock()
	defer review.MU.Unlock()
	review.Comment = &comment
	return nil
}

func (review *BuiltinUserReview[AccountID]) newProductItem(ctx context.Context, id uint64, db productItemDatabase[AccountID], form *ProductItemForm[AccountID]) (*BuiltinProductItem[AccountID], error) {
	pItem := &BuiltinProductItem[AccountID]{
		DB: db,
		FS: review.FS,
		ProductItemForm: ProductItemForm[AccountID]{
			ID: id,
		},
	}
	if err := pItem.Init(ctx); err != nil {
		return nil, err
	}
	if form != nil {
		if err := pItem.ApplyFormObject(ctx, form); err != nil {
			return nil, err
		}
	}
	return pItem, nil
}

func (review *BuiltinUserReview[AccountID]) GetProductItem(ctx context.Context) (ProductItem[AccountID], error) {
	review.MU.RLock()
	if review.ProductItem != nil {
		defer review.MU.RUnlock()
		return review.ProductItem, nil
	}
	review.MU.RUnlock()
	id, err := review.GetID(ctx)
	if err != nil {
		return nil, err
	}
	form, err := review.UserReviewForm.Clone(ctx)
	if err != nil {
		return nil, err
	}
	pItemForm := ProductItemForm[AccountID]{}
	pid, err := review.DB.GetUserReviewProductItem(ctx, &form, id, &pItemForm, review.FS)
	if err != nil {
		return nil, err
	}
	if err := review.ApplyFormObject(ctx, &form); err != nil {
		return nil, err
	}
	pItem, err := review.newProductItem(ctx, pid, review.DB, &pItemForm)
	if err != nil {
		return nil, err
	}
	review.MU.Lock()
	defer review.MU.Unlock()
	review.ProductItem = pItem
	return pItem, nil
}

func (review *BuiltinUserReview[AccountID]) Init(ctx context.Context) error {
	return nil
}

func (review *BuiltinUserReview[AccountID]) Pulse(ctx context.Context) error {
	return nil
}

func (review *BuiltinUserReview[AccountID]) ToBuiltinObject(ctx context.Context) (*BuiltinUserReview[AccountID], error) {
	return review, nil
}

func (review *BuiltinUserReview[AccountID]) ToFormObject(ctx context.Context) (*UserReviewForm[AccountID], error) {
	review.MU.RLock()
	defer review.MU.RUnlock()
	return &review.UserReviewForm, nil
}

func (review *BuiltinUserReview[AccountID]) ApplyFormObject(ctx context.Context, form *UserReviewForm[AccountID]) error {
	review.MU.Lock()
	defer review.MU.Unlock()
	// Conditional copy: only update non-zero IDs and non-nil pointers
	if form.ID != 0 {
		review.ID = form.ID
	}
	// Check if UserAccountID is zero value (requires generic type comparison)
	var zeroAccountID AccountID
	if form.UserAccountID != zeroAccountID {
		review.UserAccountID = form.UserAccountID
	}
	if form.RatingValue != nil {
		review.RatingValue = form.RatingValue
	}
	if form.Comment != nil {
		review.Comment = form.Comment
	}
	if form.ProductItem != nil {
		review.ProductItem = form.ProductItem
	}
	return nil
}

func (form *UserReviewForm[AccountID]) Clone(ctx context.Context) (UserReviewForm[AccountID], error) {
	var cloned UserReviewForm[AccountID] = *form
	return cloned, nil
}
