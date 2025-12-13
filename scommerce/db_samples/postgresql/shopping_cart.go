package dbsamples

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MobinYengejehi/scommerce/scommerce"

	"github.com/jackc/pgx/v5/pgtype"
)

var _ scommerce.DBUserShoppingCartManager[UserAccountID] = &PostgreDatabase{}
var _ scommerce.DBUserShoppingCart[UserAccountID] = &PostgreDatabase{}

func (db *PostgreDatabase) CalculateUserShoppingCartDept(ctx context.Context, form *scommerce.UserShoppingCartForm[UserAccountID], sid uint64, shippingMethod uint64) (float64, error) {
	var dept float64
	err := db.PgxPool.QueryRow(
		ctx,
		`
			select
				coalesce((
					select
						sum(sci.quantity * pi.price)
					from shopping_cart_items sci
					join product_items pi on sci."product_item_id" = pi."id"
					where sci."cart_id" = sc."id"
				), 0)
				+ coalesce(sm.price, 0) as "dept"
			from shopping_carts sc
			left join shipping_methods sm on sm.id = $1
			where sc."id" = $2
			limit 1
		`,
		shippingMethod,
		sid,
	).Scan(&dept)
	if err != nil {
		return 0, err
	}
	if form != nil {
		form.Dept = &dept
	}
	return dept, nil
}

func (db *PostgreDatabase) GetUserShoppingCartItemCount(ctx context.Context, form *scommerce.UserShoppingCartForm[UserAccountID], sid uint64) (uint64, error) {
	var count uint64
	err := db.PgxPool.QueryRow(
		ctx,
		`select count("id") from shopping_cart_items where "cart_id" = $1`,
		sid,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	if form != nil {
		form.ShoppingCartItemCount = &count
	}
	return count, nil
}

func (db *PostgreDatabase) GetUserShoppingCartItems(ctx context.Context, form *scommerce.UserShoppingCartForm[UserAccountID], sid uint64, items []uint64, itemForms []*scommerce.UserShoppingCartItemForm[UserAccountID], skip int64, limit int64, queueOrder scommerce.QueueOrder, fs scommerce.FileStorage, osm scommerce.OrderStatusManager) ([]uint64, []*scommerce.UserShoppingCartItemForm[UserAccountID], error) {
	ids := items
	if ids == nil {
		ids = make([]uint64, 0, 10)
	}
	forms := itemForms
	if forms == nil {
		forms = make([]*scommerce.UserShoppingCartItemForm[UserAccountID], 0, cap(ids))
	}

	rows, err := db.PgxPool.Query(
		ctx,
		`
			select
				"id",
				(
					select
						sc."user_id"
					from shopping_carts sc
					where sc."id" = sci."cart_id"
					limit 1
				) as "user_id",
				"product_item_id",
				"quantity",
				coalesce((
					select
						sci."quantity" * pi."price"
					from product_items pi
					where sci."product_item_id" = pi."id"
					limit 1
				), 0) as "dept",
				coalesce("attributes", 'null'::jsonb) as "attributes"
			from shopping_cart_items sci
			where sci."cart_id" = $1
			order by "id" `+queueOrder.String()+`
			offset $2
			limit $3
		`,
		sid,
		skip,
		limit,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id uint64
		var userID UserAccountID
		var productItemID uint64
		var quantity int64
		var dept float64
		var attrs json.RawMessage
		if err := rows.Scan(&id, &userID, &productItemID, &quantity, &dept, &attrs); err != nil {
			return nil, nil, err
		}

		var item *scommerce.BuiltinProductItem[UserAccountID] = nil
		if productItemID != 0 {
			item = &scommerce.BuiltinProductItem[UserAccountID]{
				DB: db,
				FS: fs,
				ProductItemForm: scommerce.ProductItemForm[UserAccountID]{
					ID: productItemID,
				},
			}
		}

		ids = append(ids, id)
		forms = append(forms, &scommerce.UserShoppingCartItemForm[UserAccountID]{
			ID:            id,
			UserAccountID: userID,
			ProductItem:   item,
			Quantity:      &quantity,
			Dept:          &dept,
			Attributes:    &attrs,
			ShoppingCart: &scommerce.BuiltinUserShoppingCart[UserAccountID]{
				DB:                 db,
				FS:                 fs,
				OrderStatusManager: osm,
				UserShoppingCartForm: scommerce.UserShoppingCartForm[UserAccountID]{
					ID:            sid,
					UserAccountID: userID,
				},
			},
		})
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return ids, forms, nil
}

func (db *PostgreDatabase) GetUserShoppingCartSessionText(ctx context.Context, form *scommerce.UserShoppingCartForm[UserAccountID], sid uint64) (string, error) {
	var sessionText pgtype.Text
	err := db.PgxPool.QueryRow(
		ctx,
		`select "session_text" from shopping_carts where "id" = $1`,
		sid,
	).Scan(&sessionText)
	if err != nil {
		return "", err
	}
	if sessionText.Valid {
		if form != nil {
			form.SessionText = &sessionText.String
		}
		return sessionText.String, nil
	}
	if form != nil {
		form.SessionText = nil
	}
	return "", nil
}

func (db *PostgreDatabase) NewUserShoppingCartShoppingCartItem(ctx context.Context, form *scommerce.UserShoppingCartForm[UserAccountID], sid uint64, productItem uint64, count int64, attrs json.RawMessage, itemForm *scommerce.UserShoppingCartItemForm[UserAccountID], fs scommerce.FileStorage, osm scommerce.OrderStatusManager) (uint64, error) {
	var id uint64
	err := db.PgxPool.QueryRow(
		ctx,
		`
			insert into
				shopping_cart_items("cart_id", "product_item_id", "quantity", "attributes")
				values($1, $2, $3, $4)
				returning "id";
		`,
		sid,
		productItem,
		count,
		attrs,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	if itemForm != nil {
		itemForm.ID = id
		itemForm.ShoppingCart = &scommerce.BuiltinUserShoppingCart[UserAccountID]{
			DB:                 db,
			FS:                 fs,
			OrderStatusManager: osm,
			UserShoppingCartForm: scommerce.UserShoppingCartForm[UserAccountID]{
				ID: sid,
			},
		}
		itemForm.ProductItem = &scommerce.BuiltinProductItem[UserAccountID]{
			DB: db,
			FS: fs,
			ProductItemForm: scommerce.ProductItemForm[UserAccountID]{
				ID: productItem,
			},
		}
		itemForm.Quantity = &count
		itemForm.Attributes = &attrs
	}
	return id, nil
}

func (db *PostgreDatabase) OrderUserShoppingCart(ctx context.Context, form *scommerce.UserShoppingCartForm[UserAccountID], sid uint64, paymentMethod uint64, address uint64, shippingMethod uint64, userComment string, orderForm *scommerce.UserOrderForm[UserAccountID]) (uint64, error) {
	var orderID uint64
	var userID UserAccountID
	var orderDate time.Time
	var orderTotal float64
	var productItemCount uint64

	err := db.PgxPool.QueryRow(
		ctx,
		`select * from order_shopping_cart($1, $2, $3, $4, $5, $6)`,
		sid,
		paymentMethod,
		address,
		shippingMethod,
		idleOrderStatusID,
		userComment,
	).Scan(&orderID, &userID, &orderDate, &orderTotal, &productItemCount)
	if err != nil {
		return 0, err
	}

	if orderForm != nil {
		orderForm.ID = orderID
		orderForm.UserAccountID = userID
		orderForm.Date = &orderDate
		orderForm.Total = &orderTotal
		orderForm.ProductItemCount = &productItemCount
		orderForm.PaymentMethod = &scommerce.BuiltinUserPaymentMethod[UserAccountID]{
			DB: db,
			UserPaymentMethodForm: scommerce.UserPaymentMethodForm[UserAccountID]{
				ID: paymentMethod,
			},
		}
		orderForm.ShippingAddress = &scommerce.BuiltinUserAddress[UserAccountID]{
			DB: db,
			UserAddressForm: scommerce.UserAddressForm[UserAccountID]{
				ID: address,
			},
		}
		orderForm.ShippingMethod = &scommerce.BuiltinShippingMethod{
			DB: db,
			ShippingMethodForm: scommerce.ShippingMethodForm{
				ID: shippingMethod,
			},
		}
		orderForm.Status = &scommerce.BuiltinOrderStatus{
			DB: db,
			OrderStatusForm: scommerce.OrderStatusForm{
				ID: idleOrderStatusID,
			},
		}
	}

	return orderID, nil
}

func (db *PostgreDatabase) RemoveUserShoppingCartAllShoppingCartItems(ctx context.Context, form *scommerce.UserShoppingCartForm[UserAccountID], sid uint64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`delete from shopping_cart_items where "cart_id" = $1`,
		sid,
	)
	if err != nil {
		return err
	}
	return nil
}

func (db *PostgreDatabase) RemoveUserShoppingCartShoppingCartItem(ctx context.Context, form *scommerce.UserShoppingCartForm[UserAccountID], sid uint64, itid uint64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`delete from shopping_cart_items where "id" = $1 and "cart_id" = $2`,
		itid,
		sid,
	)
	if err != nil {
		return err
	}
	return nil
}

func (db *PostgreDatabase) SetUserShoppingCartSessionText(ctx context.Context, form *scommerce.UserShoppingCartForm[UserAccountID], sid uint64, text string) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update shopping_carts set "session_text" = $1 where "id" = $2`,
		text,
		sid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.SessionText = &text
	}
	return nil
}

func (db *PostgreDatabase) GetShoppingCartBySessionText(ctx context.Context, sessionText string, cartForm *scommerce.UserShoppingCartForm[UserAccountID]) (uint64, error) {
	var id uint64
	var userID UserAccountID
	var dept float64
	err := db.PgxPool.QueryRow(
		ctx,
		`
			select
				sc."id",
				sc."user_id",
				coalesce((
					select
						sum(sci.quantity * pi.price)
					from shopping_cart_items sci
					join product_items pi on sci."product_item_id" = pi."id"
					where sci."cart_id" = sc."id"
				), 0) as "dept"
			from shopping_carts sc
			where sc."session_text" = $1
			limit 1
		`,
		sessionText,
	).Scan(&id, &userID, &dept)
	if err != nil {
		return 0, err
	}
	if cartForm != nil {
		cartForm.ID = id
		cartForm.UserAccountID = userID
		cartForm.SessionText = &sessionText
		cartForm.Dept = &dept
	}
	return id, nil
}

func (db *PostgreDatabase) GetShoppingCartCount(ctx context.Context) (uint64, error) {
	var count uint64
	err := db.PgxPool.QueryRow(
		ctx,
		`select count("id") from shopping_carts`,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (db *PostgreDatabase) GetShoppingCarts(ctx context.Context, carts []scommerce.DBUserUserShoppingCartResult[UserAccountID], cartForms []*scommerce.UserShoppingCartForm[UserAccountID], skip int64, limit int64, order scommerce.QueueOrder) ([]scommerce.DBUserUserShoppingCartResult[UserAccountID], []*scommerce.UserShoppingCartForm[UserAccountID], error) {
	ids := carts
	if ids == nil {
		ids = make([]scommerce.DBUserUserShoppingCartResult[UserAccountID], 0, 10)
	}
	forms := cartForms
	if forms == nil {
		forms = make([]*scommerce.UserShoppingCartForm[UserAccountID], 0, cap(ids))
	}

	rows, err := db.PgxPool.Query(
		ctx,
		`
			select
				sc."id",
				sc."user_id",
				sc."session_text",
				coalesce((
					select
						sum(sci.quantity * pi.price)
					from shopping_cart_items sci
					join product_items pi on sci."product_item_id" = pi."id"
					where sci."cart_id" = sc."id"
				), 0) as "dept"
			from shopping_carts sc
			order by "id" `+order.String()+`
			offset $1
			limit $2
		`,
		skip,
		limit,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id uint64
		var userID UserAccountID
		var sessionText pgtype.Text
		var dept float64
		if err := rows.Scan(&id, &userID, &sessionText, &dept); err != nil {
			return nil, nil, err
		}
		ids = append(ids, scommerce.DBUserUserShoppingCartResult[UserAccountID]{
			ID:  id,
			AID: userID,
		})
		form := &scommerce.UserShoppingCartForm[UserAccountID]{
			ID:            id,
			UserAccountID: userID,
			SessionText:   nil,
			Dept:          &dept,
		}
		if sessionText.Valid {
			form.SessionText = &sessionText.String
		}
		forms = append(forms, form)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return ids, forms, nil
}

func (db *PostgreDatabase) InitUserShoppingCartManager(ctx context.Context) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`
			create table if not exists shopping_carts(
				id           bigint generated by default as identity primary key,
				user_id      bigint references users(id),
				session_text text unique
			);

			create table if not exists shopping_cart_items(
				id              bigint generated by default as identity primary key,
				cart_id         bigint not null references shopping_carts(id) on delete cascade,
				product_item_id bigint not null references product_items(id),
				quantity        bigint not null,
				attributes      jsonb,
				unique (cart_id, product_item_id)
			);

			create or replace function order_shopping_cart(
				cart_id_arg bigint,
				payment_method_arg bigint,
				address_arg bigint,
				shipping_method_arg bigint,
				idle_status_id_arg bigint,
				user_comment_arg text default null
			) returns table(
				order_id bigint,
				user_id_result bigint,
				order_date_result date,
				order_total_result double precision,
				product_item_count bigint
			) as $$
			declare
				v_user_id bigint;
				v_order_id bigint;
				v_total double precision;
				v_product_items jsonb;
				v_count bigint;
			begin
				select sc.user_id into v_user_id
				from shopping_carts sc
				where sc.id = cart_id_arg;

				if v_user_id is null then
					raise exception 'Shopping cart not found';
				end if;

				select jsonb_agg(
					jsonb_build_object(
						'product_item_id', sci.product_item_id,
						'quantity', sci.quantity,
						'attributes', coalesce(sci.attributes, 'null'::jsonb)
					)
				), count(*)
				into v_product_items, v_count
				from shopping_cart_items sci
				where sci.cart_id = cart_id_arg;

				select coalesce(sum(sci.quantity * pi.price), 0) + coalesce(sm.price, 0)
				into v_total
				from shopping_cart_items sci
				join product_items pi on sci.product_item_id = pi.id
				cross join shipping_methods sm
				where sci.cart_id = cart_id_arg and sm.id = shipping_method_arg
				group by sm.price;

				insert into orders (
					user_id,
					order_date,
					payment_method_id,
					shipping_address_id,
					shipping_method_id,
					order_total,
					order_status_id,
					product_items,
					user_comment
				) values (
					v_user_id,
					current_date,
					payment_method_arg,
					address_arg,
					shipping_method_arg,
					v_total,
					idle_status_id_arg,
					v_product_items,
					user_comment_arg
				)
				returning id into v_order_id;

				delete from shopping_carts where "id" = cart_id_arg;

				return query
				select 
					v_order_id,
					v_user_id,
					current_date,
					v_total,
					v_count;
			end;
			$$ language plpgsql;
		`,
	)
	return err
}

func (db *PostgreDatabase) RemoveAllShoppingCarts(ctx context.Context) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`delete from shopping_carts`,
	)
	return err
}
