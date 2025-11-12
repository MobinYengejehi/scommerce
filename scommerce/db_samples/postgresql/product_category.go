package dbsamples

import (
	"context"
	"encoding/json"

	"github.com/MobinYengejehi/scommerce/scommerce"

	"github.com/jackc/pgx/v5/pgtype"
)

var _ scommerce.DBProductCategory[UserAccountID] = &PostgreDatabase{}

func (db *PostgreDatabase) GetProductCategoryName(ctx context.Context, form *scommerce.ProductCategoryForm[UserAccountID], pid uint64) (string, error) {
	var name string
	err := db.PgxPool.QueryRow(
		ctx,
		`select "name" from product_categories where "id" = $1 limit 1`,
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

func (db *PostgreDatabase) GetProductCategoryParent(ctx context.Context, form *scommerce.ProductCategoryForm[UserAccountID], pid uint64, catForm *scommerce.ProductCategoryForm[UserAccountID], fs scommerce.FileStorage) (uint64, error) {
	var parentID pgtype.Int8
	err := db.PgxPool.QueryRow(
		ctx,
		`select "parent_category_id" from product_categories where "id" = $1 limit 1`,
		pid,
	).Scan(&parentID)
	if err != nil {
		return 0, err
	}
	if parentID.Valid {
		if form != nil {
			form.ParentProductCategory = db.newProductCategory(parentID, fs)
		}
		return uint64(parentID.Int64), nil
	} else {
		if form != nil {
			form.ParentProductCategory = nil
		}
	}
	return 0, nil
}

func (db *PostgreDatabase) GetProductCategoryProductCount(ctx context.Context, form *scommerce.ProductCategoryForm[UserAccountID], pid uint64) (uint64, error) {
	var count uint64
	err := db.PgxPool.QueryRow(
		ctx,
		`select count("id") from products where "category_id" = $1`,
		pid,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	if form != nil {
		form.ProductCount = &count
	}
	return count, nil
}

func (db *PostgreDatabase) GetProductCategoryProducts(ctx context.Context, form *scommerce.ProductCategoryForm[UserAccountID], pid uint64, products []uint64, productForms []*scommerce.ProductForm[UserAccountID], skip int64, limit int64, queueOrder scommerce.QueueOrder, fs scommerce.FileStorage) ([]uint64, []*scommerce.ProductForm[UserAccountID], error) {
	ids := products
	if ids == nil {
		ids = make([]uint64, 0, 10)
	}
	forms := productForms
	if forms == nil {
		forms = make([]*scommerce.ProductForm[UserAccountID], 0, cap(ids))
	}

	rows, err := db.PgxPool.Query(
		ctx,
		`select "id", "name", "description", "product_images", "category_id" from products where "category_id" = $1 order by "id" `+queueOrder.String()+` offset $2 limit $3`,
		pid,
		skip,
		limit,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	cat := db.newProductCategory(pgtype.Int8{
		Valid: true,
		Int64: int64(pid),
	}, fs)

	for rows.Next() {
		var id uint64
		var name string
		var description pgtype.Text
		var productImages json.RawMessage
		var categoryID pgtype.Int8
		if err := rows.Scan(&id, &name, &description, &productImages, &categoryID); err != nil {
			return nil, nil, err
		}

		var desc *string = nil
		if description.Valid {
			desc = &description.String
		}

		var images []string
		if productImages != nil {
			if err := json.Unmarshal(productImages, &images); err != nil {
				return nil, nil, err
			}
		}

		ids = append(ids, id)
		forms = append(forms, &scommerce.ProductForm[UserAccountID]{
			ID:              id,
			Name:            &name,
			Description:     desc,
			Images:          db.getSafeImages(images),
			ProductCategory: cat,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return ids, forms, nil
}

func (db *PostgreDatabase) NewProductCategoryProduct(ctx context.Context, form *scommerce.ProductCategoryForm[UserAccountID], pid uint64, name string, description string, images []string, productForm *scommerce.ProductForm[UserAccountID], fs scommerce.FileStorage) (uint64, error) {
	var id uint64

	var jImages json.RawMessage = nil
	if images != nil {
		var err error = nil
		jImages, err = json.Marshal(images)
		if err != nil {
			return 0, err
		}
	}

	err := db.PgxPool.QueryRow(
		ctx,
		`
			insert into products(
				"name",
				"description",
				"product_images",
				"category_id"
			)
			values($1, $2, $3, $4)
			returning "id"
		`,
		name,
		description,
		jImages,
		pid,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	if productForm != nil {
		productForm.ID = id
		productForm.Name = &name
		productForm.Description = &description
		productForm.Images = &images
		productForm.ProductCategory = db.newProductCategory(pgtype.Int8{
			Valid: true,
			Int64: int64(pid),
		}, fs)
	}

	return id, nil
}

func (db *PostgreDatabase) RemoveAllProducts(ctx context.Context, form *scommerce.ProductCategoryForm[UserAccountID], pid uint64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`delete from products where "category_id" = $1`,
		pid,
	)
	return err
}

func (db *PostgreDatabase) RemoveProduct(ctx context.Context, form *scommerce.ProductCategoryForm[UserAccountID], pid uint64, product uint64) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`delete from products where "category_id" = $1 and "id" = $2`,
		pid,
		product,
	)
	return err
}

func (db *PostgreDatabase) SetProductCategoryName(ctx context.Context, form *scommerce.ProductCategoryForm[UserAccountID], pid uint64, name string) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update product_categories set "name" = $1 where "id" = $2`,
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

func (db *PostgreDatabase) SetProductCategoryParent(ctx context.Context, form *scommerce.ProductCategoryForm[UserAccountID], pid uint64, parent *uint64, fs scommerce.FileStorage) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`update product_categories set "parent_category_id" = $1 where "id" = $2`,
		parent,
		pid,
	)
	if err != nil {
		return err
	}
	if form != nil {
		if parent != nil {
			form.ParentProductCategory = db.newProductCategory(pgtype.Int8{
				Valid: true,
				Int64: int64(*parent),
			}, fs)
		} else {
			form.ParentProductCategory = nil
		}
	}
	return nil
}
