package main

import (
	"context"
	"fmt"
	"time"

	"github.com/MobinYengejehi/scommerce/scommerce"
	dbsamples "github.com/MobinYengejehi/scommerce/scommerce/db_samples/postgresql"
	"github.com/MobinYengejehi/scommerce/scommerce/fs"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	postgreConf, err := pgxpool.ParseConfig("postgresql://sika-ecommerce:sikapass123@localhost:5435/ecommerce?sslmode=disable")
	if err != nil {
		panic(err)
	}
	db, err := dbsamples.NewPostgreDatabase(ctx, postgreConf)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	lfs := fs.NewLocalDiskFileStorage("./scommerce-files")
	defer lfs.Close(ctx)
	app, err := scommerce.NewBuiltinApplication(&scommerce.AppConfig[uint64]{
		DB:             db,
		FileStorage:    lfs,
		OTPCodeLength:  8,
		OTPTokenLength: 32,
		OTPTTL:         time.Minute * 2,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("started", app)
}
