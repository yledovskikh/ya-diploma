package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"
	"github.com/yledovskikh/ya-diploma/internal/storage"
)

type DB struct {
	Pool *pgxpool.Pool
	ctx  context.Context
}

func (d *DB) Close() {
	d.Pool.Close()
}

func New(dsn string, ctx context.Context) (*DB, error) {

	dbPool, err := pgxpool.Connect(context.Background(), dsn)

	//	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{...},
	//		MaxConnections: 5,
	//		AcquireTimeout: time.Duration(10) * time.Second,
	//})

	if err != nil {
		return &DB{}, err
	}
	err = dbMigrate(dbPool, ctx)
	if err != nil {
		return &DB{}, err
	}
	return &DB{dbPool, ctx}, nil
}

func (d DB) PingDB() error {

	ctx, cancel := context.WithTimeout(d.ctx, 1*time.Second)
	defer cancel()

	if err := d.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("database is down: %w", err)
	}
	return nil
}

func dbMigrate(d *pgxpool.Pool, ctx context.Context) error {
	execSQL := []string{
		"CREATE SEQUENCE IF NOT EXISTS serial START 1",
		"CREATE TABLE IF NOT EXISTS users(id integer PRIMARY KEY DEFAULT nextval('serial'), login varchar(255) NOT NULL, password varchar(255) NOT NULL, create_at date , update_at date ,  CONSTRAINT users_unique UNIQUE (login))",
		//"CREATE TABLE IF NOT EXISTS mgauges(id integer PRIMARY KEY DEFAULT nextval('serial'), metric_name varchar(255) NOT NULL, metric_value double precision NOT NULL, CONSTRAINT mgauges_metric_name_unique UNIQUE (metric_name))",
	}

	for _, sql := range execSQL {
		_, err := d.Exec(ctx, sql)
		if err != nil {
			return fmt.Errorf("database schema was not created - %w", err)
		}
	}

	return nil
}

func (d *DB) NewUser(u storage.User) error {
	tx, err := d.Pool.Begin(d.ctx)
	if err != nil {
		return err
	}
	sql := "insert into users (login, password, create_at,update_at) values ($1,$2,$3,$4);"

	// Rollback is safe to call even if the tx is already closed, so if
	// the tx commits successfully, this is a no-op
	hp, err := HashPassword(u.Password)
	if err != nil {
		log.Error().Err(err).Msg("")
		return storage.ErrInternalServerError
	}
	defer tx.Rollback(d.ctx)
	instTime := time.Now()
	_, err = tx.Exec(d.ctx, sql, u.Login, hp, instTime, instTime)
	if err != nil {
		log.Error().Err(err).Msg("")
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return storage.ErrLoginAlreadyExist
			}
			return storage.ErrInternalServerError
		}
	}
	err = tx.Commit(d.ctx)
	if err != nil {
		log.Error().Err(err).Msg("")
		return storage.ErrInternalServerError
	}
	return nil

}

func HashPassword(password string) ([]byte, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (d *DB) CheckUser(login, password string) (bool, error) {
	return false, nil
}
