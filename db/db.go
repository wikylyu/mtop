package db

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
)

var _db *bun.DB

func ConnectToDatabase(driverName string, dsn string) (*bun.DB, error) {
	if driverName == "psql" {
		pgdb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

		return bun.NewDB(pgdb, pgdialect.New(), bun.WithDiscardUnknownColumns()), nil
	} else if driverName == "mysql" {
		sqldb, err := sql.Open("mysql", dsn)
		if err != nil {
			return nil, err
		}

		return bun.NewDB(sqldb, mysqldialect.New()), nil
	} else if driverName == "sqlite" {
		sqldb, err := sql.Open(sqliteshim.ShimName, dsn)
		if err != nil {
			return nil, err
		}

		return bun.NewDB(sqldb, sqlitedialect.New()), nil
	}
	return nil, fmt.Errorf("Unsupported database %s", driverName)
}

func Init(driverName string, dsn string, debug bool) error {
	var err error
	_db, err = ConnectToDatabase(driverName, dsn)

	if err != nil {
		return err
	}
	if debug {
		_db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	maxOpenConns := 4 * runtime.GOMAXPROCS(0)
	_db.SetMaxOpenConns(maxOpenConns)
	_db.SetMaxIdleConns(maxOpenConns)

	return nil
}

func DB() *bun.DB {
	return _db
}

func RunInTx(fn func(ctx context.Context, tx bun.Tx) error) error {
	return _db.RunInTx(context.Background(), nil, fn)
}
