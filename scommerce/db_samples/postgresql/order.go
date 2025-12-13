package dbsamples

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MobinYengejehi/scommerce/scommerce"

	"github.com/jackc/pgx/v5/pgtype"
)

var _ scommerce.DBUserOrderManager[UserAccountID] = &PostgreDatabase{}
var _ scommerce.DBUserOrder[UserAccountID] = &PostgreDatabase{}

func (db *PostgreDatabase) CalculateUserOrderTotalPrice(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64) (float64, error) {
	var total float64
	err := db.PgxPool.QueryRow(
		ctx,
		`
			select
				coalesce(sum((pi.price * (item->>'quantity')::bigint)), 0) +
				coalesce(sm.price, 0) as total
			from orders o
			left join shipping_methods sm on o.shipping_method_id = sm.id
			left join lateral jsonb_array_elements(o.product_items) as item on true
			left join product_items pi on pi.id = (item->>'product_item_id')::bigint
			where o.id = $1
			group by sm.price
		`,
		oid,
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	if form != nil {
		form.Total = &total
	}
	return total, nil
}

func (db *PostgreDatabase) DeliverUserOrder(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64, sid uint64, date time.Time, comment string) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`
			update orders
			set
				order_status_id = $1,
				delivery_date = $2,
				delivery_comment = $3
			where id = $4
		`,
		sid,
		date,
		comment,
		oid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.Status = &scommerce.BuiltinOrderStatus{
			DB: db,
			OrderStatusForm: scommerce.OrderStatusForm{
				ID: sid,
			},
		}
		form.DeliveryDate = &date
		form.DeliveryComment = &comment
	}
	return nil
}

func (db *PostgreDatabase) GetUserOrderDate(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64) (time.Time, error) {
	var date time.Time
	err := db.PgxPool.QueryRow(
		ctx,
		`select "order_date" from orders where "id" = $1`,
		oid,
	).Scan(&date)
	if err != nil {
		return time.Time{}, err
	}
	if form != nil {
		form.Date = &date
	}
	return date, nil
}

func (db *PostgreDatabase) GetUserOrderDeliveryComment(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64) (string, error) {
	var comment pgtype.Text
	err := db.PgxPool.QueryRow(
		ctx,
		`select "delivery_comment" from orders where "id" = $1`,
		oid,
	).Scan(&comment)
	if err != nil {
		return "", err
	}
	if comment.Valid {
		if form != nil {
			form.DeliveryComment = &comment.String
		}
		return comment.String, nil
	}
	if form != nil {
		form.DeliveryComment = nil
	}
	return "", nil
}

func (db *PostgreDatabase) GetUserOrderDeliveryDate(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64) (time.Time, error) {
	var date time.Time
	err := db.PgxPool.QueryRow(
		ctx,
		`select "delivery_date" from orders where "id" = $1`,
		oid,
	).Scan(&date)
	if err != nil {
		return time.Time{}, err
	}
	if form != nil {
		form.DeliveryDate = &date
	}
	return date, nil
}

func (db *PostgreDatabase) GetUserOrderPaymentMethod(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64, paymentMethodForm *scommerce.UserPaymentMethodForm[UserAccountID]) (uint64, error) {
	var id uint64
	var userID UserAccountID
	var paymentTypeID uint64
	var provider pgtype.Text
	var accountNumber pgtype.Text
	var expiryDate pgtype.Date
	var isExpired bool
	var isDefault bool
	err := db.PgxPool.QueryRow(
		ctx,
		`
			select
				pm.id,
				pm.user_id,
				pm.payment_type_id,
				pm.provider,
				pm.account_number,
				pm.expiry_date,
				pm.expiry_date < current_date as is_expired,
				pm.is_default
			from orders o
			join payment_methods pm on o.payment_method_id = pm.id
			where o.id = $1
		`,
		oid,
	).Scan(&id, &userID, &paymentTypeID, &provider, &accountNumber, &expiryDate, &isExpired, &isDefault)
	if err != nil {
		return 0, err
	}
	if paymentMethodForm != nil {
		var date *time.Time
		if expiryDate.Valid {
			date = &expiryDate.Time
		}

		var paymentType *scommerce.BuiltinPaymentType = nil
		if paymentTypeID != 0 {
			paymentType = &scommerce.BuiltinPaymentType{
				DB: db,
				PaymentTypeForm: scommerce.PaymentTypeForm{
					ID: paymentTypeID,
				},
			}
		}
		paymentMethodForm.ID = id
		paymentMethodForm.UserAccountID = userID
		paymentMethodForm.PaymentType = paymentType
		if provider.Valid {
			paymentMethodForm.Provider = &provider.String
		}
		if accountNumber.Valid {
			paymentMethodForm.AccountNumber = &accountNumber.String
		}
		paymentMethodForm.ExpiryDate = date
		paymentMethodForm.IsExpiredState = &isExpired
		paymentMethodForm.IsDefaultState = &isDefault
	}
	if form != nil {
		form.PaymentMethod = &scommerce.BuiltinUserPaymentMethod[UserAccountID]{
			DB: db,
			UserPaymentMethodForm: scommerce.UserPaymentMethodForm[UserAccountID]{
				ID:             id,
				UserAccountID:  userID,
				PaymentType:    nil,
				Provider:       nil,
				AccountNumber:  nil,
				ExpiryDate:     nil,
				IsExpiredState: &isExpired,
				IsDefaultState: &isDefault,
			},
		}
		if provider.Valid {
			form.PaymentMethod.Provider = &provider.String
		}
		if accountNumber.Valid {
			form.PaymentMethod.AccountNumber = &accountNumber.String
		}
	}
	return id, nil
}

func (db *PostgreDatabase) GetUserOrderProductItemCount(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64) (uint64, error) {
	var count uint64
	err := db.PgxPool.QueryRow(
		ctx,
		`
			select
				jsonb_array_length(product_items)
			from orders
			where id = $1
		`,
		oid,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	if form != nil {
		form.ProductItemCount = &count
	}
	return count, nil
}

func (db *PostgreDatabase) GetUserOrderProductItems(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64, items []scommerce.DBUserOrderProductItem, skip int64, limit int64, queueOrder scommerce.QueueOrder) ([]scommerce.DBUserOrderProductItem, error) {
	itms := items
	if itms == nil {
		itms = make([]scommerce.DBUserOrderProductItem, 0, 10)
	}

	rows, err := db.PgxPool.Query(
		ctx,
		`
			select
				(item->>'product_item_id')::bigint as product_item_id,
				(item->>'quantity')::bigint as quantity,
				coalesce(item->'attributes', 'null'::jsonb) as attributes
			from orders o
			cross join lateral jsonb_array_elements(o.product_items) as item
			where o.id = $1
			order by (item->>'product_item_id')::bigint `+queueOrder.String()+`
			offset $2
			limit $3
		`,
		oid,
		skip,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var productItemID uint64
		var quantity uint64
		var attrs json.RawMessage
		if err := rows.Scan(&productItemID, &quantity, &attrs); err != nil {
			return nil, err
		}

		itms = append(itms, scommerce.DBUserOrderProductItem{
			ProductItemID: productItemID,
			Quantity:      quantity,
			Attributes:    attrs,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return itms, nil
}

func (db *PostgreDatabase) GetUserOrderShippingAddress(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64, addressForm *scommerce.UserAddressForm[UserAccountID]) (uint64, error) {
	var id uint64
	var userID UserAccountID
	var unitNumber *string
	var streetNumber *string
	var addressLine1 *string
	var addressLine2 *string
	var city *string
	var region *string
	var postalCode *string
	var countryID *uint64
	var isDefault bool
	err := db.PgxPool.QueryRow(
		ctx,
		`
			select
				a.id,
				a.user_id,
				a.unit_number,
				a.street_number,
				a.address_line1,
				a.address_line2,
				a.city,
				a.region,
				a.postal_code,
				a.country_id,
				a.is_default
			from orders o
			join addresses a on o.shipping_address_id = a.id
			where o.id = $1
		`,
		oid,
	).Scan(&id, &userID, &unitNumber, &streetNumber, &addressLine1, &addressLine2, &city, &region, &postalCode, &countryID, &isDefault)
	if err != nil {
		return 0, err
	}
	if addressForm != nil {
		var country *scommerce.BuiltinCountry = nil
		if countryID != nil && *countryID != 0 {
			country = &scommerce.BuiltinCountry{
				DB: db,
				CountryForm: scommerce.CountryForm{
					ID: *countryID,
				},
			}
		}
		addressForm.ID = id
		addressForm.UserAccountID = userID
		addressForm.UnitNumber = unitNumber
		addressForm.StreetNumber = streetNumber
		addressForm.AddressLine1 = addressLine1
		addressForm.AddressLine2 = addressLine2
		addressForm.City = city
		addressForm.Region = region
		addressForm.PostalCode = postalCode
		addressForm.Country = country
		addressForm.IsDefaultState = &isDefault
	}
	if form != nil {
		form.ShippingAddress = &scommerce.BuiltinUserAddress[UserAccountID]{
			DB: db,
			UserAddressForm: scommerce.UserAddressForm[UserAccountID]{
				ID:             id,
				UserAccountID:  userID,
				UnitNumber:     unitNumber,
				StreetNumber:   streetNumber,
				AddressLine1:   addressLine1,
				AddressLine2:   addressLine2,
				City:           city,
				Region:         region,
				PostalCode:     postalCode,
				Country:        nil, // Will be set when needed
				IsDefaultState: &isDefault,
			},
		}
	}
	return id, nil
}

func (db *PostgreDatabase) GetUserOrderShippingMethod(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64, shippingMethodForm *scommerce.ShippingMethodForm) (uint64, error) {
	var id uint64
	var name string
	var price float64
	err := db.PgxPool.QueryRow(
		ctx,
		`
			select
				sm.id,
				sm.name,
				sm.price
			from orders o
			join shipping_methods sm on o.shipping_method_id = sm.id
			where o.id = $1
		`,
		oid,
	).Scan(&id, &name, &price)
	if err != nil {
		return 0, err
	}
	if shippingMethodForm != nil {
		shippingMethodForm.ID = id
		shippingMethodForm.Name = &name
		shippingMethodForm.Price = &price
	}
	if form != nil {
		form.ShippingMethod = &scommerce.BuiltinShippingMethod{
			DB: db,
			ShippingMethodForm: scommerce.ShippingMethodForm{
				ID:    id,
				Name:  &name,
				Price: &price,
			},
		}
	}
	return id, nil
}

func (db *PostgreDatabase) GetUserOrderStatus(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64, statusForm *scommerce.OrderStatusForm) (uint64, error) {
	var id uint64
	var status string
	err := db.PgxPool.QueryRow(
		ctx,
		`
			select
				os.id,
				os.status
			from orders o
			join order_statuses os on o.order_status_id = os.id
			where o.id = $1
		`,
		oid,
	).Scan(&id, &status)
	if err != nil {
		return 0, err
	}
	if statusForm != nil {
		statusForm.ID = id
		statusForm.Name = &status
	}
	if form != nil {
		form.Status = &scommerce.BuiltinOrderStatus{
			DB: db,
			OrderStatusForm: scommerce.OrderStatusForm{
				ID:   id,
				Name: &status,
			},
		}
	}
	return id, nil
}

func (db *PostgreDatabase) GetUserOrderTotal(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64) (float64, error) {
	var total float64
	err := db.PgxPool.QueryRow(
		ctx,
		`select "order_total" from orders where "id" = $1`,
		oid,
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	if form != nil {
		form.Total = &total
	}
	return total, nil
}

func (db *PostgreDatabase) GetUserOrderUserComment(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64) (string, error) {
	var comment pgtype.Text
	err := db.PgxPool.QueryRow(
		ctx,
		`select "user_comment" from orders where "id" = $1`,
		oid,
	).Scan(&comment)
	if err != nil {
		return "", err
	}
	if comment.Valid {
		if form != nil {
			form.UserComment = &comment.String
		}
		return comment.String, nil
	}
	if form != nil {
		form.UserComment = nil
	}
	return "", nil
}

func (db *PostgreDatabase) SetUserOrderDate(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64, date time.Time) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update orders set "order_date" = $1 where "id" = $2`,
		date,
		oid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.Date = &date
	}
	return nil
}

func (db *PostgreDatabase) SetUserOrderDeliveryComment(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64, comment string) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update orders set "delivery_comment" = $1 where "id" = $2`,
		comment,
		oid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.DeliveryComment = &comment
	}
	return nil
}

func (db *PostgreDatabase) SetUserOrderDeliveryDate(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64, date time.Time) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update orders set "delivery_date" = $1 where "id" = $2`,
		date,
		oid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.DeliveryDate = &date
	}
	return nil
}

func (db *PostgreDatabase) SetUserOrderPaymentMethod(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64, method uint64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update orders set "payment_method_id" = $1 where "id" = $2`,
		method,
		oid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.PaymentMethod = &scommerce.BuiltinUserPaymentMethod[UserAccountID]{
			DB: db,
			UserPaymentMethodForm: scommerce.UserPaymentMethodForm[UserAccountID]{
				ID: method,
			},
		}
	}
	return nil
}

func (db *PostgreDatabase) SetUserOrderProductItems(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64, items []scommerce.DBUserOrderProductItem) error {
	itemsJson, err := json.Marshal(items)
	if err != nil {
		return err
	}
	_, err = db.PgxPool.Exec(
		ctx,
		`update orders set "product_items" = $1 where "id" = $2`,
		itemsJson,
		oid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		count := uint64(len(items))
		form.ProductItemCount = &count
	}
	return nil
}

func (db *PostgreDatabase) SetUserOrderShippingAddress(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64, address uint64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update orders set "shipping_address_id" = $1 where "id" = $2`,
		address,
		oid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.ShippingAddress = &scommerce.BuiltinUserAddress[UserAccountID]{
			DB: db,
			UserAddressForm: scommerce.UserAddressForm[UserAccountID]{
				ID: address,
			},
		}
	}
	return nil
}

func (db *PostgreDatabase) SetUserOrderShippingMethod(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64, method uint64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update orders set "shipping_method_id" = $1 where "id" = $2`,
		method,
		oid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.ShippingMethod = &scommerce.BuiltinShippingMethod{
			DB: db,
			ShippingMethodForm: scommerce.ShippingMethodForm{
				ID: method,
			},
		}
	}
	return nil
}

func (db *PostgreDatabase) SetUserOrderStatus(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64, status uint64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update orders set "order_status_id" = $1 where "id" = $2`,
		status,
		oid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.Status = &scommerce.BuiltinOrderStatus{
			DB: db,
			OrderStatusForm: scommerce.OrderStatusForm{
				ID: status,
			},
		}
	}
	return nil
}

func (db *PostgreDatabase) SetUserOrderTotal(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64, total float64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update orders set "order_total" = $1 where "id" = $2`,
		total,
		oid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.Total = &total
	}
	return nil
}

func (db *PostgreDatabase) SetUserOrderUserComment(ctx context.Context, form *scommerce.UserOrderForm[UserAccountID], oid uint64, comment string) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update orders set "user_comment" = $1 where "id" = $2`,
		comment,
		oid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.UserComment = &comment
	}
	return nil
}

func (db *PostgreDatabase) GetUserOrderCount(ctx context.Context) (uint64, error) {
	var count uint64
	err := db.PgxPool.QueryRow(
		ctx,
		`select count("id") from orders`,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (db *PostgreDatabase) GetUserOrders(ctx context.Context, orders []scommerce.DBUserOrderResult[UserAccountID], orderForms []*scommerce.UserOrderForm[UserAccountID], skip int64, limit int64, queueOrder scommerce.QueueOrder) ([]scommerce.DBUserOrderResult[UserAccountID], []*scommerce.UserOrderForm[UserAccountID], error) {
	ids := orders
	if ids == nil {
		ids = make([]scommerce.DBUserOrderResult[UserAccountID], 0, 10)
	}
	forms := orderForms
	if forms == nil {
		forms = make([]*scommerce.UserOrderForm[UserAccountID], 0, cap(ids))
	}

	rows, err := db.PgxPool.Query(
		ctx,
		`
			select
				o.id,
				o.user_id,
				o.order_date,
				o.order_total,
				o.user_comment,
				o.delivery_date,
				o.delivery_comment,
				o.product_items,
				pm.id as payment_method_id,
				sm.id as shipping_method_id,
				os.id as order_status_id,
				a.id as shipping_address_id
			from orders o
			left join payment_methods pm on o.payment_method_id = pm.id
			left join shipping_methods sm on o.shipping_method_id = sm.id
			left join order_statuses os on o.order_status_id = os.id
			left join addresses a on o.shipping_address_id = a.id
			order by o.id `+queueOrder.String()+`
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
		var orderDate time.Time
		var orderTotal float64
		var userComment *string
		var deliveryDate *time.Time
		var deliveryComment *string
		var productItems []byte
		var paymentMethodID *uint64
		var shippingMethodID *uint64
		var orderStatusID *uint64
		var shippingAddressID *uint64

		if err := rows.Scan(
			&id,
			&userID,
			&orderDate,
			&orderTotal,
			&userComment,
			&deliveryDate,
			&deliveryComment,
			&productItems,
			&paymentMethodID,
			&shippingMethodID,
			&orderStatusID,
			&shippingAddressID,
		); err != nil {
			return nil, nil, err
		}

		ids = append(ids, scommerce.DBUserOrderResult[UserAccountID]{
			ID:  id,
			AID: userID,
		})

		form := &scommerce.UserOrderForm[UserAccountID]{
			ID:               id,
			UserAccountID:    userID,
			Date:             &orderDate,
			Total:            &orderTotal,
			UserComment:      userComment,
			DeliveryDate:     deliveryDate,
			DeliveryComment:  deliveryComment,
			ProductItemCount: nil, // Will be calculated when needed
		}

		// Set related objects if IDs are not null
		if paymentMethodID != nil {
			form.PaymentMethod = &scommerce.BuiltinUserPaymentMethod[UserAccountID]{
				DB: db,
				UserPaymentMethodForm: scommerce.UserPaymentMethodForm[UserAccountID]{
					ID: *paymentMethodID,
				},
			}
		}

		if shippingMethodID != nil {
			form.ShippingMethod = &scommerce.BuiltinShippingMethod{
				DB: db,
				ShippingMethodForm: scommerce.ShippingMethodForm{
					ID: *shippingMethodID,
				},
			}
		}

		if orderStatusID != nil {
			form.Status = &scommerce.BuiltinOrderStatus{
				DB: db,
				OrderStatusForm: scommerce.OrderStatusForm{
					ID: *orderStatusID,
				},
			}
		}

		if shippingAddressID != nil {
			form.ShippingAddress = &scommerce.BuiltinUserAddress[UserAccountID]{
				DB: db,
				UserAddressForm: scommerce.UserAddressForm[UserAccountID]{
					ID: *shippingAddressID,
				},
			}
		}

		forms = append(forms, form)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return ids, forms, nil
}

func (db *PostgreDatabase) InitUserOrderManager(ctx context.Context) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`
			create table if not exists orders(
				id                  bigint generated by default as identity primary key,
				user_id             bigint references users(id),
				order_date          date not null default current_date,
				payment_method_id   bigint references payment_methods(id),
				shipping_address_id bigint references addresses(id),
				shipping_method_id  bigint references shipping_methods(id),
				order_total         double precision not null default 0,
				order_status_id     bigint references order_statuses(id),
				user_comment        text,
				delivery_date       timestamptz,
				delivery_comment    text,
				product_items       jsonb
			);
		`,
	)
	return err
}

func (db *PostgreDatabase) RemoveAllUserOrders(ctx context.Context) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`delete from orders`,
	)
	return err
}
