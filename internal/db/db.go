package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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
		"CREATE TABLE IF NOT EXISTS users(id integer PRIMARY KEY DEFAULT nextval('users_serial'), login varchar(255) NOT NULL, password varchar(255) NOT NULL, create_at timestamp , update_at timestamp ,  CONSTRAINT users_unique UNIQUE (login))",
		"CREATE TABLE IF NOT EXISTS orders(id bigint not null primary key, user_id integer not null, status varchar(19) not null, accrual real not null, created_at timestamp not null, updated_at timestamp not null);",
	}

	for _, sql := range execSQL {
		_, err := d.Exec(ctx, sql)
		if err != nil {
			return fmt.Errorf("database schema was not created - %w", err)
		}
	}

	return nil
}

func (d *DB) NewUser(u storage.User) (int, error) {
	//tx, err := d.Pool.Begin(d.ctx)
	//if err != nil {
	//	return err
	//}
	sql := "insert into users (login, password, create_at,update_at) values ($1,$2,$3,$4) returning id;"

	// Rollback is safe to call even if the tx is already closed, so if
	// the tx commits successfully, this is a no-op
	hp, err := HashPassword(u.Password)
	if err != nil {
		log.Error().Err(err).Msg("")
		return 0, storage.ErrInternalServerError
	}
	//defer tx.Rollback(d.ctx)
	instTime := time.Now()
	var userID int
	err = d.Pool.QueryRow(d.ctx, sql, u.Login, hp, instTime, instTime).Scan(&userID)

	//if err != nil {
	//	log.Fatal(err)
	//}

	//defer stmt.Close()var studentID int
	//err = stmt.QueryRow("Lee", "Provoost").Scan(&studentId)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//_, err = d.Pool.Exec(d.ctx, sql, u.Login, hp, instTime, instTime)
	if err != nil {
		log.Error().Err(err).Msg("")
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return 0, storage.ErrLoginAlreadyExist
			}
			return 0, storage.ErrInternalServerError
		}
	}
	//err = tx.Commit(d.ctx)
	//if err != nil {
	//	log.Error().Err(err).Msg("")
	//	return storage.ErrInternalServerError
	//}
	return userID, nil

}

func HashPassword(password string) ([]byte, error) {
	hp, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return nil, err
	}
	return hp, nil
}

func (d *DB) CheckUser(u storage.User) (int, error) {
	sql := "select id, password from users where login=$1;"
	var (
		userID   int
		password string
	)

	err := d.Pool.QueryRow(d.ctx, sql, u.Login).Scan(&userID, &password)
	if err != nil {
		log.Error().Err(err).Msg("")
		if err == pgx.ErrNoRows {
			return 0, storage.ErrUnauthorized
		}
		return 0, storage.ErrInternalServerError
	}
	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(u.Password))
	if err != nil {
		return 0, storage.ErrUnauthorized
	}
	return userID, nil
}
func (d *DB) SetOrder(userID int, orderNumber int) error {
	//sql := "select id from users where login=$1;"
	//var userID int
	//err := d.Pool.QueryRow(d.ctx, sql, login).Scan(&userID)
	//if err != nil {
	//	log.Error().Err(err).Msg("")
	//	return storage.ErrInternalServerError
	//}

	instTime := time.Now()
	//_, err = d.Pool.Exec(d.ctx, sql, u.Login, hp, instTime, instTime)
	sql := "INSERT INTO orders (id, user_id, status, accrual,created_at,updated_at) VALUES($1, $2,$3,$4,$5,$6);"
	_, err := d.Pool.Exec(d.ctx, sql, orderNumber, userID, "NEW", 0, instTime, instTime)
	if err != nil {
		log.Error().Err(err).Msg("")
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return d.handleErrLoadOrder(userID, orderNumber)
			}
		}
		return storage.ErrInternalServerError
	}
	return nil
}

func (d *DB) handleErrLoadOrder(userID, orderNumber int) error {
	sql := "select user_id from orders where id=$1;"
	var userOrderID int
	err := d.Pool.QueryRow(d.ctx, sql, orderNumber).Scan(&userOrderID)
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

func (d *DB) GetOrders(userID int) ([]storage.Order, error) {
	//var userOrderID int
	//err := d.Pool.QueryRow(d.ctx, sql, order).Scan(&userOrderID)
	//if err != nil {
	//	log.Error().Err(err).Msg("")
	//	return storage.ErrInternalServerError
	//}

	sql := "select id,accrual,status,created_at,updated_at from orders where user_id=$1 order by created_at;"
	rows, err := d.Pool.Query(d.ctx, sql, userID)
	if err != nil {
		log.Error().Err(err).Msg("")
		return nil, storage.ErrInternalServerError
	}
	defer rows.Close()
	orders := make([]storage.Order, 0)
	for rows.Next() {
		var id int
		var accrual float32
		var status string
		var createdAt, updatedAt time.Time
		if err = rows.Scan(&id, &accrual, &status, &createdAt, &updatedAt); err != nil {
			return nil, storage.ErrInternalServerError
		}

		order := storage.Order{ID: strconv.Itoa(id), Status: status, Accrual: accrual, CreateAt: createdAt.Format(time.RFC3339)}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		log.Error().Err(err).Msg("")
		return nil, storage.ErrInternalServerError
	}

	return orders, nil
}
