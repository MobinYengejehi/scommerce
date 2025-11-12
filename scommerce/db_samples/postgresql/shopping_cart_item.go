package dbsamples

import (
	"context"

	"github.com/MobinYengejehi/scommerce/scommerce"
)

var _ scommerce.DBUserShoppingCartItem[UserAccountID] = &PostgreDatabase{}

func (db *PostgreDatabase) AddUserShoppingCartItemQuantity(ctx context.Context, form *scommerce.UserShoppingCartItemForm[UserAccountID], itid uint64, delta int64) error {
	var newQuantity int64
	err := db.PgxPool.QueryRow(
		ctx,
		`update shopping_cart_items set "quantity" = "quantity" + $1 where "id" = $2 returning "quantity"`,
		delta,
		itid,
	).Scan(&newQuantity)
	if err != nil {
		return err
	}
	if form != nil {
		form.Quantity = &newQuantity
	}
	return nil
}

func (db *PostgreDatabase) CalculateUserShoppingCartItemDept(ctx context.Context, form *scommerce.UserShoppingCartItemForm[UserAccountID], itid uint64) (float64, error) {
	var dept float64
	err := db.PgxPool.QueryRow(
		ctx,
		`
			select
				coalesce((
					select
						sci."quantity" * pi."price"
					from product_items pi
					where sci."product_item_id" = pi."id"
					limit 1
				), 0) as "dept"
			from shopping_cart_items sci
			where sci."id" = $1
			limit 1
		`,
		itid,
	).Scan(&dept)
	if err != nil {
		return 0, err
	}
	if form != nil {
		form.Dept = &dept
	}
	return dept, nil
}

func (db *PostgreDatabase) GetUserShoppingCartItemProductItem(ctx context.Context, form *scommerce.UserShoppingCartItemForm[UserAccountID], itid uint64, pItemForm *scommerce.ProductItemForm[UserAccountID], fs scommerce.FileStorage) (uint64, error) {
	var id uint64
	err := db.PgxPool.QueryRow(
		ctx,
		`select "product_item_id" from shopping_cart_items where "id" = $1 limit 1`,
		itid,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	if form != nil {
		form.ProductItem = &scommerce.BuiltinProductItem[UserAccountID]{
			DB: db,
			FS: fs,
			ProductItemForm: scommerce.ProductItemForm[UserAccountID]{
				ID: id,
			},
		}
	}
	return id, nil
}

func (db *PostgreDatabase) GetUserShoppingCartItemQuantity(ctx context.Context, form *scommerce.UserShoppingCartItemForm[UserAccountID], itid uint64) (int64, error) {
	var quantity int64
	err := db.PgxPool.QueryRow(
		ctx,
		`select "quantity" from shopping_cart_items where "id" = $1 limit 1`,
		itid,
	).Scan(&quantity)
	if err != nil {
		return 0, err
	}
	if form != nil {
		form.Quantity = &quantity
	}
	return quantity, nil
}

func (db *PostgreDatabase) GetUserShoppingCartItemShoppingCart(ctx context.Context, form *scommerce.UserShoppingCartItemForm[UserAccountID], itid uint64, cartForm *scommerce.UserShoppingCartForm[UserAccountID], fs scommerce.FileStorage, osm scommerce.OrderStatusManager) (uint64, error) {
	var id uint64
	err := db.PgxPool.QueryRow(
		ctx,
		`select "cart_id" from shopping_cart_items where "id" = $1 limit 1`,
		itid,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	if form != nil {
		form.ShoppingCart = &scommerce.BuiltinUserShoppingCart[UserAccountID]{
			DB:                 db,
			FS:                 fs,
			OrderStatusManager: osm,
			UserShoppingCartForm: scommerce.UserShoppingCartForm[UserAccountID]{
				ID: id,
			},
		}
	}
	return id, nil
}

func (db *PostgreDatabase) SetUserShoppingCartItemQuantity(ctx context.Context, form *scommerce.UserShoppingCartItemForm[UserAccountID], itid uint64, quantity int64) error {
	var newQuantity int64
	err := db.PgxPool.QueryRow(
		ctx,
		`update shopping_cart_items set "quantity" = $1 where "id" = $2 returning "quantity"`,
		quantity,
		itid,
	).Scan(&newQuantity)
	if err != nil {
		return err
	}
	if form != nil {
		form.Quantity = &quantity
	}
	return nil
}
