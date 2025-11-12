package dbsamples

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MobinYengejehi/scommerce/scommerce"

	"github.com/jackc/pgx/v5/pgtype"
)

var _ scommerce.DBUserAccount[UserAccountID] = &PostgreDatabase{}

func getStringPtr(text pgtype.Text) *string {
	if text.Valid {
		return &text.String
	}
	return nil
}

func (db *PostgreDatabase) AllowUserAccountTrading(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, state bool) error {
	return db.SetUserAccountActive(ctx, form, aid, state)
}

func (db *PostgreDatabase) BanUserAccount(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, till time.Duration, reason string) error {
	banTill := time.Now().Add(till)
	_, err := db.PgxPool.Exec(
		ctx,
		`update users set "ban_till" = $1, "ban_reason" = $2 where "id" = $3`,
		banTill,
		reason,
		aid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.IsBannedState = &reason
	}
	return nil
}

func (db *PostgreDatabase) CalculateUserAccountTotalDepts(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) (float64, error) {
	var total float64
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
				), 0) + coalesce(
					case
						when u.wallet < 0 then
							abs(u.wallet)
						else
							0
					end
				, 0) as "dept"
			from shopping_carts sc
			left join users u on u."id" = sc."user_id"
			where sc."user_id" = $1
			limit 1
		`,
		aid,
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	if form != nil {
		form.TotalDepts = &total
	}
	return total, nil
}

func (db *PostgreDatabase) CalculateUserAccountTotalDeptsWithoutPenalty(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) (float64, error) {
	var total float64
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
				), 0) as "dept"
			from shopping_carts sc
			where sc."user_id" = $1
			limit 1
		`,
		aid,
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	if form != nil {
		form.TotalDeptsWithoutPenalty = &total
	}
	return total, nil
}

func (db *PostgreDatabase) ChargeUserAccountWallet(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, currency float64) error {
	var newWallet float64
	err := db.PgxPool.QueryRow(
		ctx,
		`update users set "wallet" = "wallet" + $1 where "id" = $2 returning "wallet"`,
		currency,
		aid,
	).Scan(&newWallet)
	if err != nil {
		return err
	}
	if form != nil {
		form.WalletCurrency = &newWallet
	}
	return nil
}

func (db *PostgreDatabase) FineUserAccount(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, amount float64) error {
	var newPenalty float64
	err := db.PgxPool.QueryRow(
		ctx,
		`update users set "wallet" = "wallet" - $1 where "id" = $2 returning "wallet"`,
		amount,
		aid,
	).Scan(&newPenalty)
	if err != nil {
		return err
	}
	if form != nil {
		form.Penalty = &newPenalty
	}
	return nil
}

func (db *PostgreDatabase) GetUserAccountAddressCount(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) (uint64, error) {
	var count uint64
	err := db.PgxPool.QueryRow(
		ctx,
		`select count("id") from addresses where "user_id" = $1`,
		aid,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	if form != nil {
		form.AddressCount = &count
	}
	return count, nil
}

func (db *PostgreDatabase) GetUserAccountAddresses(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, addresses []uint64, addressForms []*scommerce.UserAddressForm[UserAccountID], skip int64, limit int64, queueOrder scommerce.QueueOrder) ([]uint64, []*scommerce.UserAddressForm[UserAccountID], error) {
	ids := addresses
	if ids == nil {
		ids = make([]uint64, 0, 10)
	}
	forms := addressForms
	if forms == nil {
		forms = make([]*scommerce.UserAddressForm[UserAccountID], 0, cap(ids))
	}

	rows, err := db.PgxPool.Query(
		ctx,
		`
			select
				"id",
				"unit_number",
				"street_number",
				"address_line1",
				"address_line2",
				"city",
				"region",
				"postal_code",
				"country_id",
				"is_default"
			from addresses
			where "user_id" = $1
			order by "id" `+queueOrder.String()+`
			offset $2
			limit $3
		`,
		aid,
		skip,
		limit,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id uint64
		var unitNumber pgtype.Text
		var streetNumber pgtype.Text
		var addressLine1 pgtype.Text
		var addressLine2 pgtype.Text
		var city pgtype.Text
		var region pgtype.Text
		var postalCode pgtype.Text
		var countryID pgtype.Int8
		var isDefault bool
		if err := rows.Scan(
			&id,
			&unitNumber,
			&streetNumber,
			&addressLine1,
			&addressLine2,
			&city,
			&region,
			&postalCode,
			&countryID,
			&isDefault,
		); err != nil {
			return nil, nil, err
		}

		aform := &scommerce.UserAddressForm[UserAccountID]{
			ID:             id,
			UserAccountID:  aid,
			UnitNumber:     getStringPtr(unitNumber),
			StreetNumber:   getStringPtr(streetNumber),
			AddressLine1:   getStringPtr(addressLine1),
			AddressLine2:   getStringPtr(addressLine2),
			City:           getStringPtr(city),
			Region:         getStringPtr(region),
			PostalCode:     getStringPtr(postalCode),
			IsDefaultState: &isDefault,
		}
		if countryID.Valid {
			aform.Country = &scommerce.BuiltinCountry{
				DB: db,
				CountryForm: scommerce.CountryForm{
					ID: uint64(countryID.Int64),
				},
			}
		}

		ids = append(ids, id)
		forms = append(forms, aform)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return ids, forms, nil
}

func (db *PostgreDatabase) GetUserAccountBio(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) (string, error) {
	var bio pgtype.Text
	err := db.PgxPool.QueryRow(
		ctx,
		`select "bio" from users where "id" = $1`,
		aid,
	).Scan(&bio)
	if err != nil {
		return "", err
	}
	if bio.Valid {
		if form != nil {
			form.Bio = &bio.String
		}
		return bio.String, nil
	}
	if form != nil {
		form.Bio = nil
	}
	return "", nil
}

func (db *PostgreDatabase) GetUserAccountDefaultAddress(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, addressForm *scommerce.UserAddressForm[UserAccountID]) (uint64, error) {
	var id uint64
	var unitNumber pgtype.Text
	var streetNumber pgtype.Text
	var addressLine1 pgtype.Text
	var addressLine2 pgtype.Text
	var city pgtype.Text
	var region pgtype.Text
	var postalCode pgtype.Text
	var countryID pgtype.Int8
	var isDefault bool
	err := db.PgxPool.QueryRow(
		ctx,
		`
			(
				select
					"id",
					"unit_number",
					"street_number",
					"address_line1",
					"address_line2",
					"city",
					"region",
					"postal_code",
					"country_id",
					"is_default"
				from addresses
				where "user_id" = $1 and "is_default" = true
				limit 1
			)
			union all
			(
				select
					"id",
					"unit_number",
					"street_number",
					"address_line1",
					"address_line2",
					"city",
					"region",
					"postal_code",
					"country_id",
					"is_default"
				from addresses
				where user_id = $1
				order by "id" desc
				limit 1
			)
			limit 1
		`,
		aid,
	).Scan(
		&id,
		&unitNumber,
		&streetNumber,
		&addressLine1,
		&addressLine2,
		&city,
		&region,
		&postalCode,
		&countryID,
		&isDefault,
	)
	if err != nil {
		return 0, err
	}
	if addressForm != nil {
		addressForm.ID = id
		addressForm.UserAccountID = aid
		addressForm.UnitNumber = getStringPtr(unitNumber)
		addressForm.StreetNumber = getStringPtr(streetNumber)
		addressForm.AddressLine1 = getStringPtr(addressLine1)
		addressForm.AddressLine2 = getStringPtr(addressLine2)
		addressForm.City = getStringPtr(city)
		addressForm.Region = getStringPtr(region)
		addressForm.PostalCode = getStringPtr(postalCode)
		addressForm.IsDefaultState = &isDefault
		if countryID.Valid {
			addressForm.Country = &scommerce.BuiltinCountry{
				DB: db,
				CountryForm: scommerce.CountryForm{
					ID: uint64(countryID.Int64),
				},
			}
		}
	}
	if form != nil {
		form.DefaultAddress = &scommerce.BuiltinUserAddress[UserAccountID]{
			DB: db,
			UserAddressForm: scommerce.UserAddressForm[UserAccountID]{
				ID:             id,
				UserAccountID:  aid,
				UnitNumber:     getStringPtr(unitNumber),
				StreetNumber:   getStringPtr(streetNumber),
				AddressLine1:   getStringPtr(addressLine1),
				AddressLine2:   getStringPtr(addressLine2),
				City:           getStringPtr(city),
				Region:         getStringPtr(region),
				PostalCode:     getStringPtr(postalCode),
				IsDefaultState: &isDefault,
			},
		}
	}
	return id, nil
}

func (db *PostgreDatabase) GetUserAccountDefaultPaymentMethod(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, paymentMethodForm *scommerce.UserPaymentMethodForm[UserAccountID]) (uint64, error) {
	var id uint64
	var userID uint64
	var paymentTypeID uint64
	var provider pgtype.Text
	var accountNumber pgtype.Text
	var expiryDate pgtype.Date
	var isExpired bool
	var isDefault bool
	err := db.PgxPool.QueryRow(
		ctx,
		`
			(
				select
					"id",
					"user_id",
					"payment_type_id",
					"provider",
					"account_number",
					"expiry_date",
					"expiry_date" < current_date as "is_expired",
					"is_default"
				from payment_methods
				where "user_id" = $1 and "is_default" = true
				limit 1
			)
			union all
			(
				select
					"id",
					"user_id",
					"payment_type_id",
					"provider",
					"account_number",
					"expiry_date",
					"expiry_date" < current_date as "is_expired",
					"is_default"
				from payment_methods
				where "user_id" = $1
				order by "id" desc
				limit 1
			)
			limit 1
		`,
		aid,
	).Scan(
		&id,
		&userID,
		&paymentTypeID,
		&provider,
		&accountNumber,
		&expiryDate,
		&isExpired,
		&isDefault,
	)
	if err != nil {
		return 0, err
	}
	if paymentMethodForm != nil {
		paymentMethodForm.ID = id
		paymentMethodForm.UserAccountID = userID
		if paymentTypeID != 0 {
			paymentMethodForm.PaymentType = &scommerce.BuiltinPaymentType{
				PaymentTypeForm: scommerce.PaymentTypeForm{
					ID: paymentTypeID,
				},
				DB: db,
			}
		}
		paymentMethodForm.Provider = getStringPtr(provider)
		paymentMethodForm.AccountNumber = getStringPtr(accountNumber)
		if expiryDate.Valid {
			paymentMethodForm.ExpiryDate = &expiryDate.Time
		}
		paymentMethodForm.IsExpiredState = &isExpired
		paymentMethodForm.IsDefaultState = &isDefault
	}
	if form != nil {
		form.DefaultPaymentMethod = &scommerce.BuiltinUserPaymentMethod[UserAccountID]{
			DB: db,
			UserPaymentMethodForm: scommerce.UserPaymentMethodForm[UserAccountID]{
				ID:             id,
				UserAccountID:  userID,
				PaymentType:    nil, // Will be set when needed
				Provider:       getStringPtr(provider),
				AccountNumber:  getStringPtr(accountNumber),
				ExpiryDate:     nil, // Will be set when needed
				IsExpiredState: &isExpired,
				IsDefaultState: &isDefault,
			},
		}
	}
	return id, nil
}

func (db *PostgreDatabase) GetUserAccountFirstName(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) (string, error) {
	var firstName pgtype.Text
	err := db.PgxPool.QueryRow(
		ctx,
		`select "first_name" from users where "id" = $1`,
		aid,
	).Scan(&firstName)
	if err != nil {
		return "", err
	}
	if firstName.Valid {
		if form != nil {
			form.FirstName = &firstName.String
		}
		return firstName.String, nil
	}
	if form != nil {
		form.FirstName = nil
	}
	return "", nil
}

func (db *PostgreDatabase) GetUserAccountLastName(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) (string, error) {
	var lastName pgtype.Text
	err := db.PgxPool.QueryRow(
		ctx,
		`select "last_name" from users where "id" = $1`,
		aid,
	).Scan(&lastName)
	if err != nil {
		return "", err
	}
	if lastName.Valid {
		if form != nil {
			form.LastName = &lastName.String
		}
		return lastName.String, nil
	}
	if form != nil {
		form.LastName = nil
	}
	return "", nil
}

func (db *PostgreDatabase) GetUserAccountLastUpdatedAt(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) (time.Time, error) {
	var updatedAt time.Time
	err := db.PgxPool.QueryRow(
		ctx,
		`select "updated_at" from users where "id" = $1`,
		aid,
	).Scan(&updatedAt)
	if err != nil {
		return time.Time{}, err
	}
	if form != nil {
		form.LastUpdatedAt = &updatedAt
	}
	return updatedAt, nil
}

func (db *PostgreDatabase) GetUserAccountLevel(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) (int64, error) {
	var level int64
	err := db.PgxPool.QueryRow(
		ctx,
		`select "level" from users where "id" = $1`,
		aid,
	).Scan(&level)
	if err != nil {
		return 0, err
	}
	if form != nil {
		form.UserLevel = &level
	}
	return level, nil
}

func (db *PostgreDatabase) GetUserAccountOrderCount(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) (uint64, error) {
	var count uint64
	err := db.PgxPool.QueryRow(
		ctx,
		`select count("id") from orders where "user_id" = $1`,
		aid,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	if form != nil {
		form.OrderCount = &count
	}
	return count, nil
}

func (db *PostgreDatabase) GetUserAccountOrders(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, orders []uint64, orderForms []*scommerce.UserOrderForm[UserAccountID], skip int64, limit int64, queueOrder scommerce.QueueOrder) ([]uint64, []*scommerce.UserOrderForm[UserAccountID], error) {
	ids := orders
	if ids == nil {
		ids = make([]uint64, 0, 10)
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
			where o.user_id = $1
			order by o.id `+queueOrder.String()+`
			offset $2
			limit $3
		`,
		aid,
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

		ids = append(ids, id)

		orderForm := &scommerce.UserOrderForm[UserAccountID]{
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
			orderForm.PaymentMethod = &scommerce.BuiltinUserPaymentMethod[UserAccountID]{
				DB: db,
				UserPaymentMethodForm: scommerce.UserPaymentMethodForm[UserAccountID]{
					ID: *paymentMethodID,
				},
			}
		}

		if shippingMethodID != nil {
			orderForm.ShippingMethod = &scommerce.BuiltinShippingMethod{
				DB: db,
				ShippingMethodForm: scommerce.ShippingMethodForm{
					ID: *shippingMethodID,
				},
			}
		}

		if orderStatusID != nil {
			orderForm.Status = &scommerce.BuiltinOrderStatus{
				DB: db,
				OrderStatusForm: scommerce.OrderStatusForm{
					ID: *orderStatusID,
				},
			}
		}

		if shippingAddressID != nil {
			orderForm.ShippingAddress = &scommerce.BuiltinUserAddress[UserAccountID]{
				DB: db,
				UserAddressForm: scommerce.UserAddressForm[UserAccountID]{
					ID: *shippingAddressID,
				},
			}
		}

		forms = append(forms, orderForm)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return ids, forms, nil
}

func (db *PostgreDatabase) GetUserAccountPassword(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) (string, error) {
	var password string
	err := db.PgxPool.QueryRow(
		ctx,
		`select "password" from users where "id" = $1`,
		aid,
	).Scan(&password)
	if err != nil {
		return "", err
	}
	if form != nil {
		form.Password = &password
	}
	return password, nil
}

func (db *PostgreDatabase) GetUserAccountPaymentMethodCount(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) (uint64, error) {
	var count uint64
	err := db.PgxPool.QueryRow(
		ctx,
		`select count("id") from payment_methods where "user_id" = $1`,
		aid,
	).Scan(&count)
	if err != nil {
		return 0, nil
	}
	if form != nil {
		form.PaymentMethodCount = &count
	}
	return count, nil
}

func (db *PostgreDatabase) GetUserAccountPaymentMethods(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, paymentMethods []uint64, paymentMethodForms []*scommerce.UserPaymentMethodForm[UserAccountID], skip int64, limit int64, queueOrder scommerce.QueueOrder) ([]uint64, []*scommerce.UserPaymentMethodForm[UserAccountID], error) {
	ids := paymentMethods
	if ids == nil {
		ids = make([]uint64, 0, 10)
	}
	forms := paymentMethodForms
	if forms == nil {
		forms = make([]*scommerce.UserPaymentMethodForm[UserAccountID], 0, cap(ids))
	}

	rows, err := db.PgxPool.Query(
		ctx,
		`
			select
				"id",
				"user_id",
				"payment_type_id",
				"provider",
				"account_number",
				"expiry_date",
			 	"expiry_date" < current_date as "is_expired",
				"is_default"
			from payment_methods
			where "user_id" = $1
			order by "id" `+queueOrder.String()+`
			offset $2
			limit $3
		`,
		aid,
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
		var paymentTypeID uint64
		var provider pgtype.Text
		var accountNumber pgtype.Text
		var expiryDate pgtype.Date
		var isExpired bool
		var isDefault bool
		if err := rows.Scan(
			&id,
			&userID,
			&paymentTypeID,
			&provider,
			&accountNumber,
			&expiryDate,
			&isExpired,
			&isDefault,
		); err != nil {
			return nil, nil, err
		}

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

		ids = append(ids, id)
		forms = append(forms, &scommerce.UserPaymentMethodForm[UserAccountID]{
			ID:             id,
			UserAccountID:  userID,
			PaymentType:    paymentType,
			Provider:       getStringPtr(provider),
			AccountNumber:  getStringPtr(accountNumber),
			ExpiryDate:     date,
			IsExpiredState: &isExpired,
			IsDefaultState: &isDefault,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return ids, forms, nil
}

func (db *PostgreDatabase) GetUserAccountProfileImages(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) ([]string, error) {
	var imagesRaw json.RawMessage
	err := db.PgxPool.QueryRow(
		ctx,
		`select "profile_images" from users where "id" = $1`,
		aid,
	).Scan(&imagesRaw)
	if err != nil {
		return nil, err
	}

	var images []string
	if imagesRaw != nil {
		if err := json.Unmarshal(imagesRaw, &images); err != nil {
			return nil, err
		}
	}

	if form != nil {
		form.ProfileImages = db.getSafeImages(images)
	}

	return images, nil
}

func (db *PostgreDatabase) GetUserAccountRole(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, roleForm *scommerce.UserRoleForm) (uint64, error) {
	var roleID pgtype.Int8
	err := db.PgxPool.QueryRow(
		ctx,
		`select "role_id" from users where "id" = $1`,
		aid,
	).Scan(&roleID)
	if err != nil {
		return 0, err
	}

	if roleID.Valid {
		if roleForm != nil {
			roleForm.ID = uint64(roleID.Int64)
		}
		if form != nil {
			form.Role = &scommerce.BuiltinUserRole{
				DB: db,
				UserRoleForm: scommerce.UserRoleForm{
					ID: uint64(roleID.Int64),
				},
			}
		}
		return uint64(roleID.Int64), nil
	}

	return 0, nil
}

func (db *PostgreDatabase) GetUserAccountShoppingCartCount(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) (uint64, error) {
	var count uint64
	err := db.PgxPool.QueryRow(
		ctx,
		`select count("id") from shopping_carts where "user_id" = $1`,
		aid,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	if form != nil {
		form.ShoppingCartCount = &count
	}
	return count, nil
}

func (db *PostgreDatabase) GetUserAccountShoppingCarts(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, carts []uint64, cartForms []*scommerce.UserShoppingCartForm[UserAccountID], skip int64, limit int64, queueOrder scommerce.QueueOrder) ([]uint64, []*scommerce.UserShoppingCartForm[UserAccountID], error) {
	ids := carts
	if ids == nil {
		ids = make([]uint64, 0, 10)
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
				sc."session_text",
				coalesce((
					select
						sum(sci.quantity * pi.price)
					from shopping_cart_items sci
					join product_items pi on sci."product_item_id" = pi."id"
					where sci."cart_id" = sc."id"
				), 0) as "dept"
			from shopping_carts sc
			where "user_id" = $1
			order by "id" `+queueOrder.String()+`
			offset $2
			limit $3
		`,
		aid,
		skip,
		limit,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id uint64
		var sessionText string
		var dept float64
		if err := rows.Scan(&id, &sessionText, &dept); err != nil {
			return nil, nil, err
		}
		ids = append(ids, id)
		forms = append(forms, &scommerce.UserShoppingCartForm[UserAccountID]{
			ID:            id,
			UserAccountID: aid,
			SessionText:   &sessionText,
			Dept:          &dept,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return ids, forms, nil

}

func (db *PostgreDatabase) GetUserAccountToken(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) (string, error) {
	var token string
	err := db.PgxPool.QueryRow(
		ctx,
		`select "token" from users where "id" = $1`,
		aid,
	).Scan(&token)
	if err != nil {
		return "", err
	}
	if form != nil {
		form.Token = &token
	}
	return token, nil
}

func (db *PostgreDatabase) GetUserAccountWalletCurrency(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) (float64, error) {
	var wallet float64
	err := db.PgxPool.QueryRow(
		ctx,
		`select "wallet" from users where "id" = $1`,
		aid,
	).Scan(&wallet)
	if err != nil {
		return 0, err
	}
	if form != nil {
		form.WalletCurrency = &wallet
	}
	return wallet, nil
}

func (db *PostgreDatabase) HasUserAccountPenalty(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) (bool, error) {
	var walletCurrency float64
	err := db.PgxPool.QueryRow(
		ctx,
		`select coalesce("wallet", 0) from users where "id" = $1`,
		aid,
	).Scan(&walletCurrency)
	if err != nil {
		return false, err
	}
	hasPenalty := walletCurrency < 0
	if form != nil {
		penalty := -walletCurrency
		if hasPenalty {
			form.Penalty = &penalty
		} else {
			penalty = 0
			form.Penalty = &penalty
		}
	}
	return hasPenalty, nil
}

func (db *PostgreDatabase) IsUserAccountActive(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) (bool, error) {
	var isActive bool
	err := db.PgxPool.QueryRow(
		ctx,
		`select "is_active" from users where "id" = $1`,
		aid,
	).Scan(&isActive)
	if err != nil {
		return false, err
	}
	if form != nil {
		form.IsActiveState = &isActive
	}
	return isActive, nil
}

func (db *PostgreDatabase) IsUserAccountBanned(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) (string, error) {
	var banTill pgtype.Timestamptz
	var banReason pgtype.Text
	err := db.PgxPool.QueryRow(
		ctx,
		`select "ban_till", "ban_reason" from users where "id" = $1`,
		aid,
	).Scan(&banTill, &banReason)
	if err != nil {
		return "", err
	}

	if banTill.Valid && banTill.Time.After(time.Now()) && banReason.Valid {
		if form != nil {
			form.IsBannedState = &banReason.String
		}
		return banReason.String, nil
	}

	if form != nil {
		form.IsBannedState = nil
	}

	return "", nil
}

func (db *PostgreDatabase) IsUserAccountSuperUser(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) (bool, error) {
	var isSuperUser bool
	err := db.PgxPool.QueryRow(
		ctx,
		`select "level" > 0 as "is_superuser" from users where "id" = $1`,
		aid,
	).Scan(&isSuperUser)
	if err != nil {
		return false, err
	}
	if form != nil {
		form.IsSuperUserState = &isSuperUser
	}
	return isSuperUser, nil
}

func (db *PostgreDatabase) IsUserAccountTradingAllowed(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) (bool, error) {
	var isTradingAllowed bool
	err := db.PgxPool.QueryRow(
		ctx,
		`select "is_active" from users where "id" = $1`,
		aid,
	).Scan(&isTradingAllowed)
	if err != nil {
		return false, err
	}
	if form != nil {
		form.IsTradingAllowedState = &isTradingAllowed
	}
	return isTradingAllowed, nil
}

func (db *PostgreDatabase) NewUserAccountAddress(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, unitNumber string, street_number string, addressLine1 string, addressLine2 string, city string, region string, postalCode string, country *uint64, isDefault bool, addressForm *scommerce.UserAddressForm[UserAccountID]) (uint64, error) {
	var id uint64
	err := db.PgxPool.QueryRow(
		ctx,
		`
			insert into addresses(
				"user_id",
				"unit_number",
				"street_number",
				"address_line1",
				"address_line2",
				"city",
				"region",
				"postal_code",
				"country_id",
				"is_default"
			) values(
				$1,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8,
				$9,
				$10
			) returning "id"
		`,
		aid,
		unitNumber,
		street_number,
		addressLine1,
		addressLine2,
		city,
		region,
		postalCode,
		country,
		isDefault,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	if addressForm != nil {
		addressForm.ID = id
		addressForm.UserAccountID = aid
		addressForm.UnitNumber = &unitNumber
		addressForm.AddressLine1 = &addressLine1
		addressForm.AddressLine2 = &addressLine2
		addressForm.City = &city
		addressForm.Region = &region
		addressForm.PostalCode = &postalCode
		if country != nil {
			addressForm.Country = &scommerce.BuiltinCountry{
				DB: db,
				CountryForm: scommerce.CountryForm{
					ID: *country,
				},
			}
		}
		addressForm.IsDefaultState = &isDefault
	}
	return id, nil
}

func (db *PostgreDatabase) NewUserAccountPaymentMethod(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, paymentType uint64, provider string, accoutNumber string, expiryDate time.Time, isDefault bool, paymentMethodForm *scommerce.UserPaymentMethodForm[UserAccountID]) (uint64, error) {
	var id uint64
	err := db.PgxPool.QueryRow(
		ctx,
		`
			insert into payment_methods(
				"user_id",
				"payment_type_id",
				"provider",
				"account_number",
				"expiry_date",
				"is_default"
			) values(
				$1,
				$2,
				$3,
				$4,
				$5,
				$6
			) returning "id"
		`,
		aid,
		paymentType,
		provider,
		accoutNumber,
		pgtype.Date{
			Valid: true,
			Time:  expiryDate,
		},
		isDefault,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	if paymentMethodForm != nil {
		paymentMethodForm.ID = id
		paymentMethodForm.UserAccountID = aid
		paymentMethodForm.PaymentType = &scommerce.BuiltinPaymentType{
			PaymentTypeForm: scommerce.PaymentTypeForm{
				ID: paymentType,
			},
			DB: db,
		}
		paymentMethodForm.AccountNumber = &accoutNumber
		paymentMethodForm.ExpiryDate = &expiryDate
		paymentMethodForm.IsDefaultState = &isDefault
	}
	return id, nil
}

func (db *PostgreDatabase) NewUserAccountShoppingCart(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, sessionText string, cartForm *scommerce.UserShoppingCartForm[UserAccountID]) (uint64, error) {
	var id uint64
	err := db.PgxPool.QueryRow(
		ctx,
		`
			insert into
				shopping_carts("user_id", "session_text")
				values($1, $2)
				returning "id"
		`,
		aid,
		sessionText,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	if cartForm != nil {
		cartForm.ID = id
		cartForm.UserAccountID = aid
		cartForm.SessionText = &sessionText
	}
	return id, nil
}

func (db *PostgreDatabase) RemoveAllUserAccountAddresses(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`delete from addresses where "user_id" = $1`,
		aid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		count := uint64(0)
		form.AddressCount = &count
	}
	return nil
}

func (db *PostgreDatabase) RemoveAllUserAccountOrders(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`delete from orders where "user_id" = $1`,
		aid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		count := uint64(0)
		form.OrderCount = &count
	}
	return nil
}

func (db *PostgreDatabase) RemoveAllUserAccountPaymentMethods(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`delete from payment_methods where "user_id" = $1`,
		aid,
	)
	return err
}

func (db *PostgreDatabase) RemoveAllUserAccountShoppingCarts(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`delete from shopping_carts where "user_id" = $1`,
		aid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		count := uint64(0)
		form.ShoppingCartCount = &count
	}
	return nil
}

func (db *PostgreDatabase) RemoveUserAccountAddress(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, addrID uint64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`delete from addresses where "user_id" = $1 and "id" = $2`,
		aid,
		addrID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (db *PostgreDatabase) RemoveUserAccountOrder(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, oid uint64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`delete from orders where "user_id" = $1 and "id" = $2`,
		aid,
		oid,
	)
	if err != nil {
		return err
	}
	return nil
}

func (db *PostgreDatabase) RemoveUserAccountPaymentMethod(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, pid uint64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`delete from payment_methods where "user_id" = $1 and "id" = $2`,
		aid,
		pid,
	)
	return err
}

func (db *PostgreDatabase) RemoveUserAccountShoppingCart(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, cid uint64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`delete from shopping_carts where "user_id" = $1 and "id" = $2`,
		aid,
		cid,
	)
	return err
}

func (db *PostgreDatabase) SetUserAccountActive(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, state bool) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update users set "is_active" = $1 where "id" = $2`,
		state,
		aid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.IsActiveState = &state
	}
	return nil
}

func (db *PostgreDatabase) SetUserAccountBio(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, bio string) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update users set "bio" = $1 where "id" = $2`,
		bio,
		aid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.Bio = &bio
	}
	return nil
}

func (db *PostgreDatabase) SetUserAccountFirstName(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, name string) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update users set "first_name" = $1 where "id" = $2`,
		name,
		aid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.FirstName = &name
	}
	return nil
}

func (db *PostgreDatabase) SetUserAccountLastName(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, name string) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update users set "last_name" = $1 where "id" = $2`,
		name,
		aid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.LastName = &name
	}
	return nil
}

func (db *PostgreDatabase) SetUserAccountLastUpdatedAt(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, lastUpdatedAt time.Time) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update users set "updated_at" = $1 where "id" = $2`,
		lastUpdatedAt,
		aid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.LastUpdatedAt = &lastUpdatedAt
	}
	return nil
}

func (db *PostgreDatabase) SetUserAccountLevel(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, level int64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update users set "level" = $1 where "id" = $2`,
		level,
		aid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.UserLevel = &level
	}
	return nil
}

func (db *PostgreDatabase) SetUserAccountPassword(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, password string) error {
	hashedPassword, err := db.hashPassword(password)
	if err != nil {
		return err
	}
	_, err = db.PgxPool.Exec(
		ctx,
		`update users set "password" = $1 where "id" = $2`,
		hashedPassword,
		aid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.Password = &hashedPassword
	}
	return nil
}

func (db *PostgreDatabase) SetUserAccountPenalty(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, penalty float64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update users set "wallet" = -$1 where "id" = $2`,
		penalty,
		aid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.Penalty = &penalty
	}
	return nil
}

func (db *PostgreDatabase) SetUserAccountProfileImages(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, images []string) error {
	var imagesRaw json.RawMessage = nil
	if images != nil {
		var err error = nil
		imagesRaw, err = json.Marshal(images)
		if err != nil {
			return err
		}
	}
	_, err := db.PgxPool.Exec(
		ctx,
		`update users set "profile_images" = $1 where "id" = $2`,
		imagesRaw,
		aid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.ProfileImages = db.getSafeImages(images)
	}
	return nil
}

func (db *PostgreDatabase) SetUserAccountRole(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, role uint64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update users set "role_id" = $1 where "id" = $2`,
		role,
		aid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.Role = &scommerce.BuiltinUserRole{
			DB: db,
			UserRoleForm: scommerce.UserRoleForm{
				ID: role,
			},
		}
	}
	return nil
}

func (db *PostgreDatabase) SetUserAccountSuperUser(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, state bool) error {
	level := 0
	if state {
		level = 1
	}
	_, err := db.PgxPool.Exec(
		ctx,
		`update users set "level" = $1 where "id" = $2`,
		level,
		aid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.IsSuperUserState = &state
	}
	return nil
}

func (db *PostgreDatabase) SetUserAccountToken(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, token string) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update users set "token" = $1 where "id" = $2`,
		token,
		aid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.Token = &token
	}
	return nil
}

func (db *PostgreDatabase) SetUserAccountWalletCurrency(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, currency float64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update users set "wallet" = $1 where "id" = $2`,
		currency,
		aid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.WalletCurrency = &currency
	}
	return nil
}

func (db *PostgreDatabase) TransferUserAccountCurrency(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, to UserAccountID, currency float64) error {
	var sourceNewWallet float64
	var destNewWallet float64

	err := db.PgxPool.QueryRow(
		ctx,
		`select * from transfer_user_currency($1, $2, $3)`,
		aid,
		to,
		currency,
	).Scan(&sourceNewWallet, &destNewWallet)
	if err != nil {
		return err
	}

	if form != nil {
		form.WalletCurrency = &sourceNewWallet
	}

	return nil
}

func (db *PostgreDatabase) UnbanUserAccount(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update users set "ban_till" = null, "ban_reason" = null where "id" = $1`,
		aid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.IsBannedState = nil
	}
	return nil
}

func (db *PostgreDatabase) ValidateUserAccountPassword(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, password string) (bool, error) {
	var hashedPassword string
	err := db.PgxPool.QueryRow(
		ctx,
		`select "password" from users where "id" = $1`,
		aid,
	).Scan(&hashedPassword)
	if err != nil {
		return false, err
	}

	valid, err := db.validatePassword(hashedPassword, password)
	if err != nil {
		return false, err
	}

	return valid, nil
}

func (db *PostgreDatabase) GetUserAccountUserReviews(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID, ids []uint64, reviewForms []*scommerce.UserReviewForm[UserAccountID], skip int64, limit int64, queueOrder scommerce.QueueOrder) ([]uint64, []*scommerce.UserReviewForm[UserAccountID], error) {
	resultIDs := ids
	if resultIDs == nil {
		resultIDs = make([]uint64, 0, 10)
	}
	forms := reviewForms
	if forms == nil {
		forms = make([]*scommerce.UserReviewForm[UserAccountID], 0, cap(resultIDs))
	}

	rows, err := db.PgxPool.Query(
		ctx,
		`
			select
				"id",
				"user_id",
				"order_product_id",
				"rating_value",
				"comment"
			from user_reviews
			where "user_id" = $1
			order by "id" `+queueOrder.String()+`
			offset $2
			limit $3
		`,
		aid,
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
		var ratingValue int32
		var comment string
		if err := rows.Scan(
			&id,
			&userID,
			&productItemID,
			&ratingValue,
			&comment,
		); err != nil {
			return nil, nil, err
		}

		resultIDs = append(resultIDs, id)
		forms = append(forms, &scommerce.UserReviewForm[UserAccountID]{
			ID:            id,
			UserAccountID: userID,
			RatingValue:   &ratingValue,
			Comment:       &comment,
			ProductItem: &scommerce.BuiltinProductItem[UserAccountID]{
				DB: db,
				ProductItemForm: scommerce.ProductItemForm[UserAccountID]{
					ID: productItemID,
				},
			},
		})
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return resultIDs, forms, nil
}

func (db *PostgreDatabase) GetUserAccountUserReviewCount(ctx context.Context, form *scommerce.UserAccountForm[UserAccountID], aid UserAccountID) (uint64, error) {
	var count uint64
	err := db.PgxPool.QueryRow(
		ctx,
		`select count("id") from user_reviews where "user_id" = $1`,
		aid,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
