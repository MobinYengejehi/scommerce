package dbsamples

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/MobinYengejehi/scommerce/scommerce"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

var _ scommerce.DBUserAccountManager[UserAccountID] = &PostgreDatabase{}

func (db *PostgreDatabase) AuthenticateUserAccount(ctx context.Context, token string, password string, accountForm *scommerce.UserAccountForm[UserAccountID]) (UserAccountID, error) {
	var id UserAccountID
	var userPasswordHash pgtype.Text
	var firstName pgtype.Text
	var lastName pgtype.Text
	var updatedAt pgtype.Timestamptz
	var roleID pgtype.Int8
	var level int64
	var isActive bool
	var profileImages json.RawMessage
	var bio pgtype.Text
	var wallet float64
	var banTill pgtype.Timestamptz
	var banReason pgtype.Text
	err := db.PgxPool.QueryRow(
		ctx,
		`
			select
				"id",
				"password",
				"first_name",
				"last_name",
				"updated_at",
				"role_id",
				"level",
				"is_active",
				"profile_images",
				"bio",
				"wallet",
				"ban_till",
				"ban_reason"
			from users where "token" = $1 limit 1
		`,
		token,
	).Scan(
		&id,
		&userPasswordHash,
		&firstName,
		&lastName,
		&updatedAt,
		&roleID,
		&level,
		&isActive,
		&profileImages,
		&bio,
		&wallet,
		&banTill,
		&banReason,
	)
	if err != nil {
		return 0, err
	}

	if !isActive {
		return 0, errors.New("user account is not active")
	}

	if userPasswordHash.Valid {
		valid, err := db.validatePassword(userPasswordHash.String, password)
		if err != nil {
			return 0, err
		}
		if !valid {
			return 0, errors.New("user account password mismatched")
		}
	}

	var bReason *string = nil
	if banTill.Valid && banReason.Valid {
		if banTill.Time.After(time.Now()) {
			bReason = &banReason.String
		} else {
			_, err := db.PgxPool.Exec(
				ctx,
				`update users set "ban_till" = null, "ban_reason" = null where "id" = $1`,
				id,
			)
			if err != nil {
				return 0, err
			}
		}
	}

	if bReason != nil {
		return 0, errors.New("user account is banned because of '" + *bReason + "' till '" + banTill.Time.String() + "'")
	}

	if accountForm != nil {
		var images []string
		if profileImages != nil {
			if err := json.Unmarshal(profileImages, &images); err != nil {
				return 0, err
			}
		}

		accountForm.ID = id
		if userPasswordHash.Valid {
			accountForm.Password = &userPasswordHash.String
		}
		if firstName.Valid {
			accountForm.FirstName = &firstName.String
		}
		if lastName.Valid {
			accountForm.LastName = &lastName.String
		}
		if updatedAt.Valid {
			accountForm.LastUpdatedAt = &updatedAt.Time
		}
		if roleID.Valid {
			accountForm.Role = &scommerce.BuiltinUserRole{
				DB: db,
				UserRoleForm: scommerce.UserRoleForm{
					ID: uint64(roleID.Int64),
				},
			}
		}
		accountForm.UserLevel = &level
		accountForm.IsActiveState = &isActive
		accountForm.ProfileImages = db.getSafeImages(images)
		if bio.Valid {
			accountForm.Bio = &bio.String
		}
		accountForm.WalletCurrency = &wallet
		accountForm.IsBannedState = bReason
	}

	return id, nil
}

func (db *PostgreDatabase) GetUserAccount(ctx context.Context, token string, accountForm *scommerce.UserAccountForm[UserAccountID]) (UserAccountID, error) {
	var id UserAccountID
	var firstName pgtype.Text
	var lastName pgtype.Text
	var updatedAt pgtype.Timestamptz
	var roleID pgtype.Int8
	var level int64
	var isActive bool
	var profileImagesRaw json.RawMessage
	var bio pgtype.Text
	var wallet float64
	var banTill pgtype.Timestamptz
	var banReason pgtype.Text

	var err error = nil

	err = db.PgxPool.QueryRow(
		ctx,
		`
			select
				"id",
				"first_name",
				"last_name",
				"updated_at",
				"role_id",
				"level",
				"is_active",
				"profile_images",
				"bio",
				"wallet",
				"ban_till",
				"ban_reason"
			from users where "token" = $1 limit 1
		`,
		token,
	).Scan(
		&id,
		&firstName,
		&lastName,
		&updatedAt,
		&roleID,
		&level,
		&isActive,
		&profileImagesRaw,
		&bio,
		&wallet,
		&banTill,
		&banReason,
	)
	if err != nil {
		return 0, err
	}

	if accountForm != nil {
		var images []string
		if profileImagesRaw != nil {
			if err := json.Unmarshal(profileImagesRaw, &images); err != nil {
				return 0, err
			}
		}

		var bReason *string = nil
		if banTill.Valid && banReason.Valid {
			bReason = &banReason.String
		}

		accountForm.ID = id
		if firstName.Valid {
			accountForm.FirstName = &firstName.String
		}
		if lastName.Valid {
			accountForm.LastName = &lastName.String
		}
		if updatedAt.Valid {
			accountForm.LastUpdatedAt = &updatedAt.Time
		}
		if roleID.Valid {
			accountForm.Role = &scommerce.BuiltinUserRole{
				DB: db,
				UserRoleForm: scommerce.UserRoleForm{
					ID: uint64(roleID.Int64),
				},
			}
		}
		accountForm.UserLevel = &level
		accountForm.IsActiveState = &isActive
		accountForm.ProfileImages = db.getSafeImages(images)
		if bio.Valid {
			accountForm.Bio = &bio.String
		}
		accountForm.WalletCurrency = &wallet
		accountForm.IsBannedState = bReason
	}

	return id, nil
}

func (db *PostgreDatabase) GetUserAccountCount(ctx context.Context) (uint64, error) {
	var count uint64
	err := db.PgxPool.QueryRow(
		ctx,
		`select count("id") from users`,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (db *PostgreDatabase) GetUserAccounts(ctx context.Context, accounts []UserAccountID, accountForms []*scommerce.UserAccountForm[UserAccountID], skip int64, limit int64, queueOrder scommerce.QueueOrder) ([]UserAccountID, []*scommerce.UserAccountForm[UserAccountID], error) {
	ids := accounts
	if ids == nil {
		ids = make([]uint64, 0, 10)
	}
	forms := accountForms
	if forms != nil {
		forms = make([]*scommerce.UserAccountForm[UserAccountID], 0, cap(ids))
	}

	rows, err := db.PgxPool.Query(
		ctx,
		`
			select
				"id",
				"first_name",
				"last_name",
				"updated_at",
				"role_id",
				"level",
				"is_active",
				"profile_images",
				"bio",
				"wallet",
				"ban_till",
				"ban_reason"
			from users order by "id" `+queueOrder.String()+` offset $1 limit $2
		`,
		skip,
		limit,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id UserAccountID
		var firstName pgtype.Text
		var lastName pgtype.Text
		var updatedAt pgtype.Timestamptz
		var roleID pgtype.Int8
		var level int64
		var isActive bool
		var profileImages json.RawMessage
		var bio pgtype.Text
		var wallet float64
		var banTill pgtype.Timestamptz
		var banReason pgtype.Text
		if err := rows.Scan(
			&id,
			&firstName,
			&lastName,
			&updatedAt,
			&roleID,
			&level,
			&isActive,
			&profileImages,
			&bio,
			&wallet,
			&banTill,
			&banReason,
		); err != nil {
			return nil, nil, err
		}

		form := &scommerce.UserAccountForm[UserAccountID]{
			ID:             id,
			FirstName:      nil,
			LastName:       nil,
			UserLevel:      &level,
			IsActiveState:  &isActive,
			Bio:            nil,
			WalletCurrency: &wallet,
		}

		if firstName.Valid {
			form.FirstName = &firstName.String
		}
		if lastName.Valid {
			form.LastName = &lastName.String
		}
		if bio.Valid {
			form.Bio = &bio.String
		}

		if updatedAt.Valid {
			form.LastUpdatedAt = &updatedAt.Time
		}
		if roleID.Valid {
			form.Role = &scommerce.BuiltinUserRole{
				DB: db,
				UserRoleForm: scommerce.UserRoleForm{
					ID: uint64(roleID.Int64),
				},
			}
		}

		var images []string
		if profileImages != nil {
			if err := json.Unmarshal(profileImages, &images); err != nil {
				return nil, nil, err
			}
		}
		form.ProfileImages = db.getSafeImages(images)

		if banTill.Valid && banReason.Valid {
			form.IsBannedState = &banReason.String
		}

		ids = append(ids, id)
		forms = append(forms, form)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return ids, forms, nil
}

func (db *PostgreDatabase) InitUserAccountManager(ctx context.Context) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`
			create table if not exists users(
				id             bigint generated by default as identity primary key,
				token          varchar(256) not null unique,
				password       varchar(256) not null,
				first_name     varchar(256),
				last_name      varchar(256),
				updated_at     timestamptz not null default now(),
				role_id        bigint references roles(id),
				level          bigint not null default 0,
				is_active      boolean not null default true,
				profile_images jsonb,
				bio            varchar(2000),
				wallet         double precision default 0,
				ban_till       timestamptz default null,
				ban_reason     varchar(1000) default null
			);

			create or replace function transfer_user_currency(
				source_user_id bigint,
				destination_user_id bigint,
				amount double precision
			) returns table(
				source_new_wallet double precision,
				destination_new_wallet double precision
			) as $$
			declare
				v_source_wallet double precision;
				v_destination_wallet double precision;
			begin
				if amount <= 0 then
					raise exception 'Transfer amount must be positive';
				end if;

				select wallet into v_source_wallet
				from users
				where id = source_user_id
				for update;

				if v_source_wallet is null then
					raise exception 'Source user not found';
				end if;

				select wallet into v_destination_wallet
				from users
				where id = destination_user_id
				for update;

				if v_destination_wallet is null then
					raise exception 'Destination user not found';
				end if;

				if v_source_wallet < amount then
					raise exception 'Insufficient funds';
				end if;

				update users
				set wallet = wallet - amount
				where id = source_user_id;

				update users
				set wallet = wallet + amount
				where id = destination_user_id;

				v_source_wallet := v_source_wallet - amount;
				v_destination_wallet := v_destination_wallet + amount;

				return query
				select v_source_wallet, v_destination_wallet;
			end;
			$$ language plpgsql;
		`,
	)
	return err
}

func (db *PostgreDatabase) NewUserAccount(ctx context.Context, token string, password string, accountForm *scommerce.UserAccountForm[UserAccountID]) (UserAccountID, error) {
	var id UserAccountID
	var updatedAt pgtype.Timestamptz
	var roleID pgtype.Int8
	var level int64
	var isActive bool
	var wallet float64

	hashedPassword, err := db.hashPassword(password)
	if err != nil {
		return 0, err
	}

	err = db.PgxPool.QueryRow(
		ctx,
		`
			insert into users(
				"token",
				"password"
			) values(
				$1,
				$2
			) returning
				"id",
				"updated_at",
				"role_id",
				"level",
				"is_active",
				"wallet"
		`,
		token,
		hashedPassword,
	).Scan(&id, &updatedAt, &roleID, &level, &isActive, &wallet)
	if err != nil {
		return 0, err
	}

	if accountForm != nil {
		accountForm.ID = id
		accountForm.Token = &token
		accountForm.Password = &hashedPassword
		if updatedAt.Valid {
			accountForm.LastUpdatedAt = &updatedAt.Time
		}
		if roleID.Valid {
			accountForm.Role = &scommerce.BuiltinUserRole{
				DB: db,
				UserRoleForm: scommerce.UserRoleForm{
					ID: uint64(roleID.Int64),
				},
			}
		}
		accountForm.UserLevel = &level
		accountForm.IsActiveState = &isActive
		accountForm.WalletCurrency = &wallet
	}

	return id, nil
}

func (db *PostgreDatabase) RemoveAllUserAccounts(ctx context.Context) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`delete from users`,
	)
	return err
}

func (db *PostgreDatabase) RemoveUserAccount(ctx context.Context, aid UserAccountID) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`delete from users where "id" = $1`,
		aid,
	)
	return err
}

func (db *PostgreDatabase) RemoveUserAccountWithToken(ctx context.Context, token string, password string) error {
	_, err := db.PgxPool.Exec(
		ctx,
		`delete from users where "token" = $1`,
		token,
	)
	return err
}

func (db *PostgreDatabase) FillUserAccountWithID(ctx context.Context, aid UserAccountID, accountForm *scommerce.UserAccountForm[UserAccountID]) error {
	if accountForm == nil {
		return errors.New("account form is nil")
	}

	var token string
	var firstName pgtype.Text
	var lastName pgtype.Text
	var updatedAt pgtype.Timestamptz
	var roleID pgtype.Int8
	var level int64
	var isActive bool
	var profileImagesRaw json.RawMessage
	var bio pgtype.Text
	var wallet float64
	var banTill pgtype.Timestamptz
	var banReason pgtype.Text

	var err error = nil

	err = db.PgxPool.QueryRow(
		ctx,
		`
			select
				"token",
				"first_name",
				"last_name",
				"updated_at",
				"role_id",
				"level",
				"is_active",
				"profile_images",
				"bio",
				"wallet",
				"ban_till",
				"ban_reason"
			from users where "id" = $1 limit 1
		`,
		aid,
	).Scan(
		&token,
		&firstName,
		&lastName,
		&updatedAt,
		&roleID,
		&level,
		&isActive,
		&profileImagesRaw,
		&bio,
		&wallet,
		&banTill,
		&banReason,
	)
	if err != nil {
		return err
	}

	if accountForm != nil {
		var images []string
		if profileImagesRaw != nil {
			if err := json.Unmarshal(profileImagesRaw, &images); err != nil {
				return err
			}
		}

		var bReason *string = nil
		if banTill.Valid && banReason.Valid {
			bReason = &banReason.String
		}

		accountForm.ID = aid
		accountForm.Token = &token
		if firstName.Valid {
			accountForm.FirstName = &firstName.String
		}
		if lastName.Valid {
			accountForm.LastName = &lastName.String
		}
		if updatedAt.Valid {
			accountForm.LastUpdatedAt = &updatedAt.Time
		}
		if roleID.Valid {
			accountForm.Role = &scommerce.BuiltinUserRole{
				DB: db,
				UserRoleForm: scommerce.UserRoleForm{
					ID: uint64(roleID.Int64),
				},
			}
		}
		accountForm.UserLevel = &level
		accountForm.IsActiveState = &isActive
		accountForm.ProfileImages = db.getSafeImages(images)
		if bio.Valid {
			accountForm.Bio = &bio.String
		}
		accountForm.WalletCurrency = &wallet
		accountForm.IsBannedState = bReason
	}

	return nil
}

func (db *PostgreDatabase) hashPassword(password string) (string, error) {
	if len(password) == 0 {
		return "", nil
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func (db *PostgreDatabase) validatePassword(hashedPassword string, password string) (bool, error) {
	hashLen := len(hashedPassword)
	passLen := len(password)
	if hashLen == 0 && passLen == 0 {
		if passLen == 0 {
			return true, nil
		}
		return false, nil
	}
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
