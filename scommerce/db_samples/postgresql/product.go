package dbsamples

import (
	"context"
	"encoding/json"

	"github.com/MobinYengejehi/scommerce/scommerce"

	"github.com/jackc/pgx/v5/pgtype"
)

var _ scommerce.DBProduct[UserAccountID] = &PostgreDatabase{}

func (db *PostgreDatabase) AddProductProductItem(ctx context.Context, form *scommerce.ProductForm[UserAccountID], pid uint64, sku string, name string, price float64, quantity uint64, images []string, attrs json.RawMessage, itemForm *scommerce.ProductItemForm[UserAccountID], fs scommerce.FileStorage) (uint64, error) {
	var id uint64

	var err error = nil

	var jImages json.RawMessage = nil
	if images != nil {
		jImages, err = json.Marshal(images)
		if err != nil {
			return 0, err
		}
	}

	err = db.PgxPool.QueryRow(
		ctx,
		`
			insert into product_items(
				"sku",
				"name",
				"price",
				"quantity_in_stock",
				"attributes",
				"product_images",
				"product_id"
			) values(
				$1,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7
			) returning "id"
		`,
		sku,
		name,
		price,
		quantity,
		attrs,
		jImages,
		pid,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	if itemForm != nil {
		itemForm.ID = id
		itemForm.SKU = &sku
		itemForm.Name = &name
		itemForm.Price = &price
		itemForm.QuantityInStock = &quantity
		itemForm.Images = db.getSafeImages(images)
		itemForm.Attributes = &attrs
		itemForm.Product = &scommerce.BuiltinProduct[UserAccountID]{
			DB: db,
			FS: fs,
			ProductForm: scommerce.ProductForm[UserAccountID]{
				ID: pid,
			},
		}
	}

	return id, nil
}

func (db *PostgreDatabase) GetProductCategory(ctx context.Context, form *scommerce.ProductForm[UserAccountID], pid uint64, catForm *scommerce.ProductCategoryForm[UserAccountID], fs scommerce.FileStorage) (uint64, error) {
	var cid pgtype.Int8
	var name pgtype.Text
	var parentID pgtype.Int8
	err := db.PgxPool.QueryRow(
		ctx,
		`select pc.id as category_id, pc.name as category_name, pc.parent_category_id as parent_category_id from products p join product_categories pc on p.category_id = pc.id where p.id = $1`,
		pid,
	).Scan(&cid, &name, &parentID)
	if err != nil {
		return 0, err
	}
	if cid.Valid {
		if catForm != nil {
			catForm.ID = uint64(cid.Int64)
			if name.Valid {
				catForm.Name = &name.String
			}
			catForm.ParentProductCategory = db.newProductCategory(parentID, fs)
		}
		return uint64(cid.Int64), nil
	}
	return 0, nil
}

func (db *PostgreDatabase) GetProductDescription(ctx context.Context, form *scommerce.ProductForm[UserAccountID], pid uint64) (string, error) {
	var desc pgtype.Text
	err := db.PgxPool.QueryRow(
		ctx,
		`select "description" from products where "id" = $1 limit 1`,
		pid,
	).Scan(&desc)
	if err != nil {
		return "", err
	}
	if desc.Valid {
		if form != nil {
			form.Description = &desc.String
		}
		return desc.String, nil
	}
	return "", nil
}

func (db *PostgreDatabase) GetProductImages(ctx context.Context, form *scommerce.ProductForm[UserAccountID], pid uint64) ([]string, error) {
	var imagesRaw json.RawMessage
	err := db.PgxPool.QueryRow(
		ctx,
		`select "product_images" from products where "id" = $1 limit 1`,
		pid,
	).Scan(&imagesRaw)
	if err != nil {
		return nil, err
	}
	var images []string
	if err := json.Unmarshal(imagesRaw, &images); err != nil {
		return nil, err
	}
	if form != nil {
		form.Images = &images
	}
	return images, nil
}

func (db *PostgreDatabase) GetProductName(ctx context.Context, form *scommerce.ProductForm[UserAccountID], pid uint64) (string, error) {
	var name string
	err := db.PgxPool.QueryRow(
		ctx,
		`select "name" from products where "id" = $1 limit 1`,
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

func (db *PostgreDatabase) GetProductProductItemCount(ctx context.Context, form *scommerce.ProductForm[UserAccountID], pid uint64) (uint64, error) {
	var count uint64
	err := db.PgxPool.QueryRow(
		ctx,
		`select count("id") from product_items where "product_id" = $1`,
		pid,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	if form != nil {
		form.ProductItemCount = &count
	}
	return count, nil
}

func (db *PostgreDatabase) GetProductProductItems(ctx context.Context, form *scommerce.ProductForm[UserAccountID], pid uint64, items []uint64, itemForms []*scommerce.ProductItemForm[UserAccountID], skip int64, limit int64, queueOrder scommerce.QueueOrder, fs scommerce.FileStorage) ([]uint64, []*scommerce.ProductItemForm[UserAccountID], error) {
	ids := items
	if ids == nil {
		ids = make([]uint64, 0, 10)
	}
	forms := itemForms
	if forms == nil {
		forms = make([]*scommerce.ProductItemForm[UserAccountID], 0, cap(ids))
	}

	rows, err := db.PgxPool.Query(
		ctx,
		`
			select
				"id",
				"sku",
				"name",
				"price",
				"quantity_in_stock",
				"attributes",
				"product_images",
				"product_id"
			from product_items
			where "product_id" = $1
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
		var sku string
		var name string
		var price float64
		var quantityInStock int32
		var attributes json.RawMessage
		var productImages json.RawMessage
		var productID pgtype.Int8
		if err := rows.Scan(
			&id,
			&sku,
			&name,
			&price,
			&quantityInStock,
			&attributes,
			&productImages,
			&productID,
		); err != nil {
			return nil, nil, err
		}

		var images []string
		if err := json.Unmarshal(productImages, &images); err != nil {
			return nil, nil, err
		}

		var quantity uint64 = uint64(quantityInStock)

		var product *scommerce.BuiltinProduct[UserAccountID] = nil
		if productID.Valid {
			product = &scommerce.BuiltinProduct[UserAccountID]{
				DB: db,
				FS: fs,
				ProductForm: scommerce.ProductForm[UserAccountID]{
					ID: uint64(productID.Int64),
				},
			}
		}

		ids = append(ids, id)
		forms = append(forms, &scommerce.ProductItemForm[UserAccountID]{
			ID:              id,
			Attributes:      &attributes,
			Images:          db.getSafeImages(images),
			Price:           &price,
			Name:            &name,
			QuantityInStock: &quantity,
			SKU:             &sku,
			Product:         product,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return ids, forms, nil
}

func (db *PostgreDatabase) RemoveAllProductProductItems(ctx context.Context, form *scommerce.ProductForm[UserAccountID], pid uint64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`delete from product_items where "product_id" = $1`,
		pid,
	)
	return err
}

func (db *PostgreDatabase) RemoveProductProductItem(ctx context.Context, form *scommerce.ProductForm[UserAccountID], pid uint64, itid uint64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`delete from product_items where "product_id" = $1 and "id" = $2`,
		pid,
		itid,
	)
	return err
}

func (db *PostgreDatabase) SetProductCategory(ctx context.Context, form *scommerce.ProductForm[UserAccountID], pid uint64, category *uint64, fs scommerce.FileStorage) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update products set "category_id" = $1 where "id" = $2`,
		category,
		pid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.ProductCategory = db.newProductCategory(pgtype.Int8{
			Valid: true,
			Int64: int64(*category),
		}, fs)
	}
	return nil
}

func (db *PostgreDatabase) SetProductDescription(ctx context.Context, form *scommerce.ProductForm[UserAccountID], pid uint64, desc string) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update products set "description" = $1 where "id" = $2`,
		desc,
		pid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		form.Description = &desc
	}
	return nil
}

func (db *PostgreDatabase) SetProductImages(ctx context.Context, form *scommerce.ProductForm[UserAccountID], pid uint64, images []string) error {
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
		`update products set "product_images" = $1 where "id" = $2`,
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

func (db *PostgreDatabase) SetProductName(ctx context.Context, form *scommerce.ProductForm[UserAccountID], pid uint64, name string) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update products set "name" = $1 where "id" = $2`,
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
