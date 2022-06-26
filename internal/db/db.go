package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zerologadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
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

func New(dbURL string, ctx context.Context) (*DB, error) {

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return &DB{}, err
	}

	config.MaxConns = 15
	config.MinConns = 2
	config.ConnConfig.LogLevel = pgx.LogLevelInfo
	config.ConnConfig.Logger = zerologadapter.NewLogger(zerolog.New(zerolog.NewConsoleWriter()))

	//config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
	//	// do something with every new connection
	//}

	dbPool, err := pgxpool.ConnectConfig(context.Background(), config)

	//dbPool, err := pgxpool.Connect(context.Background(), dbURL)

	if err != nil {
		return &DB{}, err
	}

	//	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{...},
	//		MaxConnections: 5,
	//		AcquireTimeout: time.Duration(10) * time.Second,
	//})

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
		"CREATE SEQUENCE IF NOT EXISTS users_serial START 1",
		//"CREATE SEQUENCE IF NOT EXISTS order_serial START 1",
		"CREATE TABLE IF NOT EXISTS users(id integer PRIMARY KEY DEFAULT nextval('users_serial'), login varchar(255) NOT NULL, password varchar(255) NOT NULL, create_at date , update_at date ,  CONSTRAINT users_unique UNIQUE (login))",
		"CREATE TABLE IF NOT EXISTS orders(id bigint not null primary key, user_id integer not null, status varchar(19) not null, accrual real not null, created_at date not null, updated_at date not null);",
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
	//tx, err := d.Pool.Begin(d.ctx)
	//if err != nil {
	//	return err
	//}
	sql := "insert into users (login, password, create_at,update_at) values ($1,$2,$3,$4);"

	// Rollback is safe to call even if the tx is already closed, so if
	// the tx commits successfully, this is a no-op
	hp, err := HashPassword(u.Password)
	if err != nil {
		log.Error().Err(err).Msg("")
		return storage.ErrInternalServerError
	}
	//defer tx.Rollback(d.ctx)
	instTime := time.Now()
	_, err = d.Pool.Exec(d.ctx, sql, u.Login, hp, instTime, instTime)
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
	//err = tx.Commit(d.ctx)
	//if err != nil {
	//	log.Error().Err(err).Msg("")
	//	return storage.ErrInternalServerError
	//}
	return nil

}

func HashPassword(password string) ([]byte, error) {
	hp, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return nil, err
	}
	return hp, nil
}

func (d *DB) CheckUser(u storage.User) error {
	sql := "select password from users where login=$1;"
	var password string
	err := d.Pool.QueryRow(d.ctx, sql, u.Login).Scan(&password)
	if err != nil {
		log.Error().Err(err).Msg("")
		if err == pgx.ErrNoRows {
			return storage.ErrUnauthorized
		}
		return storage.ErrInternalServerError
	}
	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(u.Password))
	if err != nil {
		return storage.ErrUnauthorized
	}
	return nil
}
func (d *DB) SetOrder(login string, order int) error {
	sql := "select id from users where login=$1;"
	var userID int
	err := d.Pool.QueryRow(d.ctx, sql, login).Scan(&userID)
	if err != nil {
		log.Error().Err(err).Msg("")
		return storage.ErrInternalServerError
	}

	instTime := time.Now()
	//_, err = d.Pool.Exec(d.ctx, sql, u.Login, hp, instTime, instTime)
	sql = "INSERT INTO orders (id, user_id, status, accrual,created_at,updated_at) VALUES($1, $2,$3,$4,$5,$6);"
	_, err = d.Pool.Exec(d.ctx, sql, order, userID, "new", 0, instTime, instTime)
	if err != nil {
		log.Error().Err(err).Msg("")
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return d.handleErrLoadOrder(userID, order)
			}
		}
		return storage.ErrInternalServerError
	}
	return nil
}

func (d *DB) handleErrLoadOrder(userID, order int) error {
	sql := "select user_id from orders where id=$1;"
	var userOrderID int
	err := d.Pool.QueryRow(d.ctx, sql, order).Scan(&userOrderID)
	if err != nil {
		log.Error().Err(err).Msg("")
		return storage.ErrInternalServerError
	}
	if userID != userOrderID {
		log.Debug().Msgf("userID != userOrderID : %d != %d", userID, userOrderID)
		return storage.ErrOrderLoadedAnotherUser
	}
	return storage.ErrUserAlreadyLoadedOrder
}
