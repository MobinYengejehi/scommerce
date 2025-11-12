package dbsamples

import (
	"context"
	"encoding/json"

	"github.com/MobinYengejehi/scommerce/scommerce"

	"github.com/jackc/pgx/v5/pgtype"
)

var _ scommerce.DBProductItem[UserAccountID] = &PostgreDatabase{}

func (db *PostgreDatabase) AddProductItemQuantityInStock(ctx context.Context, form *scommerce.ProductItemForm[UserAccountID], pid uint64, delta int64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update product_items set "quantity_in_stock" = "quantity_in_stock" + $1 where "id" = $2`,
		delta,
		pid,
	)
	return err
}

func (db *PostgreDatabase) GetProductItemAttributes(ctx context.Context, form *scommerce.ProductItemForm[UserAccountID], pid uint64) (json.RawMessage, error) {
	var attrs json.RawMessage
	err := db.PgxPool.QueryRow(
		ctx,
		`select "attributes" from product_items where "id" = $1 limit 1`,
		pid,
	).Scan(&attrs)
	if err != nil {
		return nil, err
	}
	if form != nil {
		if attrs != nil {
			form.Attributes = &attrs
		}
	}
	return attrs, nil
}

func (db *PostgreDatabase) GetProductItemImages(ctx context.Context, form *scommerce.ProductItemForm[UserAccountID], pid uint64) ([]string, error) {
	var jImages json.RawMessage
	err := db.PgxPool.QueryRow(
		ctx,
		`select "product_images" from product_items where "id" = $1 limit 1`,
		pid,
	).Scan(&jImages)
	if err != nil {
		return nil, err
	}
	var images []string
	if err := json.Unmarshal(jImages, &images); err != nil {
		return nil, err
	}
	if form != nil {
		form.Images = db.getSafeImages(images)
	}
	return images, nil
}

func (db *PostgreDatabase) GetProductItemName(ctx context.Context, form *scommerce.ProductItemForm[UserAccountID], pid uint64) (string, error) {
	var name string
	err := db.PgxPool.QueryRow(
		ctx,
		`select "name" from product_items where "id" = $1 limit 1`,
		pid,
	).Scan(&name)
	if err != nil {
		return "", err
	}
	if form != nil {
		form.Name = &name
	}
	return name, nil
}

func (db *PostgreDatabase) GetProductItemPrice(ctx context.Context, form *scommerce.ProductItemForm[UserAccountID], pid uint64) (float64, error) {
	var price float64
	err := db.PgxPool.QueryRow(
		ctx,
		`select "price" from product_items where "id" = $1 limit 1`,
		pid,
	).Scan(&price)
	if err != nil {
		return 0, err
	}
	if form != nil {
		form.Price = &price
	}
	return price, nil
}

func (db *PostgreDatabase) GetProductItemProduct(ctx context.Context, form *scommerce.ProductItemForm[UserAccountID], pid uint64, productForm *scommerce.ProductForm[UserAccountID], fs scommerce.FileStorage) (uint64, error) {
	var id uint64
	var name string
	var description pgtype.Text
	var imagesRaw json.RawMessage
	var categoryID pgtype.Int8

	err := db.PgxPool.QueryRow(
		ctx,
		`
			select
				p.id as product_id,
				p.name as product_name,
				p.description as product_description,
				p.product_images as product_images,
				p.category_id as product_category_id
			from product_items pi
			join products p on pi.product_id = p.id
			where pi.id = $1
		`,
		pid,
	).Scan(&id, &name, &description, &imagesRaw, &categoryID)
	if err != nil {
		return 0, err
	}

	if productForm != nil {
		var images []string
		if imagesRaw != nil {
			if err := json.Unmarshal(imagesRaw, &images); err != nil {
				return 0, err
			}
		}

		var desc *string = nil
		if description.Valid {
			desc = &description.String
		}

		productForm.ID = id
		productForm.Name = &name
		productForm.Description = desc
		productForm.Images = &images
		productForm.ProductCategory = db.newProductCategory(categoryID, fs)

		if form != nil {
			form.Product = &scommerce.BuiltinProduct[UserAccountID]{
				DB:          db,
				FS:          fs,
				ProductForm: *productForm,
			}
		}
	}

	return id, nil
}

func (db *PostgreDatabase) GetProductItemQuantityInStock(ctx context.Context, form *scommerce.ProductItemForm[UserAccountID], pid uint64) (uint64, error) {
	var quantity uint64
	err := db.PgxPool.QueryRow(
		ctx,
		`select "quantity_in_stock" from product_items where "id" = $1 limit 1`,
		pid,
	).Scan(&quantity)
	if err != nil {
		return 0, err
	}
	if form != nil {
		form.QuantityInStock = &quantity
	}
	return quantity, nil
}

func (db *PostgreDatabase) GetProductItemSKU(ctx context.Context, form *scommerce.ProductItemForm[UserAccountID], pid uint64) (string, error) {
	var sku string
	err := db.PgxPool.QueryRow(
		ctx,
		`select "sku" from product_items where "id" = $1 limit 1`,
		pid,
	).Scan(&sku)
	if err != nil {
		return "", err
	}
	if form != nil {
		form.SKU = &sku
	}
	return sku, nil
}

func (db *PostgreDatabase) SetProductItemAttributes(ctx context.Context, form *scommerce.ProductItemForm[UserAccountID], pid uint64, attrs json.RawMessage) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update product_items set "attributes" = $1 where "id" = $2`,
		pid,
		attrs,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.Attributes = &attrs
	}
	return nil
}

func (db *PostgreDatabase) SetProductItemImages(ctx context.Context, form *scommerce.ProductItemForm[UserAccountID], pid uint64, images []string) error {
	var err error = nil

	var jImages json.RawMessage = nil
	if images != nil {
		jImages, err = json.Marshal(images)
		if err != nil {
			return err
		}
	}

	_, err = db.PgxPool.Exec(
		ctx,
		`update product_items set "product_images" = $1 where "id" = $2`,
		jImages,
		pid,
	)
	if err != nil {
		return err
	}

	if form != nil {
		form.Images = &images
	}

	return nil
}

func (db *PostgreDatabase) SetProductItemName(ctx context.Context, form *scommerce.ProductItemForm[UserAccountID], pid uint64, name string) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update product_items set "name" = $1 where "id" = $2`,
		name,
		pid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.Name = &name
	}
	return nil
}

func (db *PostgreDatabase) SetProductItemPrice(ctx context.Context, form *scommerce.ProductItemForm[UserAccountID], pid uint64, price float64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update product_items set "price" = $1 where "id" = $2`,
		price,
		pid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.Price = &price
	}
	return nil
}

func (db *PostgreDatabase) SetProductItemProduct(ctx context.Context, form *scommerce.ProductItemForm[UserAccountID], pid uint64, product *uint64, fs scommerce.FileStorage) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update product_items set "product_id" = $1 where "id" = $2`,
		product,
		pid,
	)
	if err != nil {
		return err
	}
	if product != nil {
		if form != nil {
			form.Product = &scommerce.BuiltinProduct[UserAccountID]{
				DB: db,
				FS: fs,
				ProductForm: scommerce.ProductForm[UserAccountID]{
					ID: *product,
				},
			}
		}
	}
	return nil
}

func (db *PostgreDatabase) SetProductItemQuantityInStock(ctx context.Context, form *scommerce.ProductItemForm[UserAccountID], pid uint64, quantity uint64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update product_items set "quantity_in_stock" = $1 where "id" = $2`,
		quantity,
		pid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.QuantityInStock = &quantity
	}
	return nil
}

func (db *PostgreDatabase) SetProductItemSKU(ctx context.Context, form *scommerce.ProductItemForm[UserAccountID], pid uint64, sku string) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update product_items set "sku" = $1 where "id" = $2`,
		sku,
		pid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.SKU = &sku
	}
	return nil
}

func (db *PostgreDatabase) GetProductItemUserReviews(ctx context.Context, form *scommerce.ProductItemForm[UserAccountID], pid uint64, ids []uint64, reviewForms []*scommerce.UserReviewForm[UserAccountID], skip int64, limit int64, queueOrder scommerce.QueueOrder) ([]uint64, []*scommerce.UserReviewForm[UserAccountID], error) {
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
			where "order_product_id" = $1
			order by "id" `+queueOrder.String()+`
			offset $2
			limit $3
		`,
		pid,
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
		var comment pgtype.Text
		if err := rows.Scan(
			&id,
			&userID,
			&productItemID,
			&ratingValue,
			&comment,
		); err != nil {
			return nil, nil, err
		}

		form := &scommerce.UserReviewForm[UserAccountID]{
			ID:            id,
			UserAccountID: userID,
			RatingValue:   &ratingValue,
			Comment:       nil,
			ProductItem: &scommerce.BuiltinProductItem[UserAccountID]{
				DB: db,
				ProductItemForm: scommerce.ProductItemForm[UserAccountID]{
					ID: productItemID,
				},
			},
		}
		if comment.Valid {
			form.Comment = &comment.String
		}

		resultIDs = append(resultIDs, id)
		forms = append(forms, form)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return resultIDs, forms, nil
}

func (db *PostgreDatabase) GetProductItemUserReviewCount(ctx context.Context, form *scommerce.ProductItemForm[UserAccountID], pid uint64) (uint64, error) {
	var count uint64
	err := db.PgxPool.QueryRow(
		ctx,
		`select count("id") from user_reviews where "order_product_id" = $1`,
		pid,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (db *PostgreDatabase) CalculateProductItemAverageRating(ctx context.Context, form *scommerce.ProductItemForm[UserAccountID], pid uint64) (float64, error) {
	var avgRating float64
	err := db.PgxPool.QueryRow(
		ctx,
		`select coalesce(avg("rating_value"), 0) from user_reviews where "order_product_id" = $1`,
		pid,
	).Scan(&avgRating)
	if err != nil {
		return 0, err
	}
	return avgRating, nil
}
