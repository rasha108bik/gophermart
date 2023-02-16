package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/rasha108bik/gophermart/internal/config"
	"github.com/rasha108bik/gophermart/internal/models"
	dbErr "github.com/rasha108bik/gophermart/internal/storage/errors"
)

type DBStorage struct {
	Cfg   config.Config
	DB    *sql.DB
	Queue Queue
}

func NewDBStorage(cfg config.Config) (*DBStorage, error) {
	s := &DBStorage{Cfg: cfg, Queue: NewSliceQueue()}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", cfg.DataBaseURI)
	if err != nil {
		log.Println("failed open DB:", err)
		return s, err
	}

	err = migrateUP(db)
	if err != nil {
		return nil, err
	}

	s.DB = db

	orders, err := s.GetOrdersForUpdate(ctx)
	if err != nil {
		log.Println("failed s.GetOrdersForUpdate")
		return s, err
	}

	err = s.Queue.PushFrontOrders(orders...)
	if err != nil {
		log.Println("failed Queue.PushFrontOrders")
		return s, err
	}

	return s, nil
}

func migrateUP(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Printf("postgres.WithInstance: %v\n", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"pgx", driver)
	if err != nil {
		log.Printf("migrate.NewWithDatabaseInstance: %v\n", err)
		return err
	}

	err = m.Up() // or m.Step(2) if you want to explicitly set the number of migrations to run
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal(fmt.Errorf("migrate failed: %v", err))
		return err
	}

	return nil
}

func (s *DBStorage) GetUser(ctx context.Context, search, value string) (models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	user := models.User{}

	var row *sql.Row

	switch search {
	case "id":
		query := `SELECT * FROM users WHERE id = $1`
		row = s.DB.QueryRowContext(ctx, query, value)
	case "login":
		query := `SELECT * FROM users WHERE login = $1`
		row = s.DB.QueryRowContext(ctx, query, value)
	}

	switch err := row.Scan(&user.ID, &user.Login, &user.Password, &user.Balance, &user.Withdrawn); err {
	case sql.ErrNoRows:
		return user, dbErr.ErrNotFound
	case nil:
		return user, nil
	default:
		log.Println("Failed get user:", err)
		return user, err
	}
}

func (s *DBStorage) AddUser(ctx context.Context, user models.User) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	_, err := s.GetUser(ctx, "login", user.Login)
	if err == nil {
		return 0, dbErr.ErrLoginExists
	}

	h := sha256.New()
	h.Write([]byte(user.Login + user.Password))
	hash := hex.EncodeToString(h.Sum(nil))

	_, err = s.DB.ExecContext(ctx, `INSERT INTO users (login, password, balance, withdrawn) VALUES ($1, $2, $3, $4)`, user.Login, hash, 0, 0)
	if err != nil {
		return 0, err
	}

	err = s.DB.QueryRowContext(ctx, `SELECT id FROM users WHERE login = $1`, user.Login).Scan(&user.ID)

	if err != nil {
		return 0, err
	}

	return user.ID, nil
}

func (s *DBStorage) Withdraw(ctx context.Context, user models.User, withdrawal models.Withdraw) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	result, err := s.DB.ExecContext(ctx, `UPDATE users SET balance = balance - $1, withdrawn = withdrawn + $1 WHERE id = $2 AND balance >= $1`, withdrawal.Sum, user.ID)

	if err != nil {
		log.Println("Failed withdraw:", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return dbErr.ErrNoMoney
	}

	_, err = s.DB.ExecContext(ctx, `INSERT INTO withdrawals (user_id, number, amount, processed) VALUES ($1, $2, $3, $4)`, user.ID, withdrawal.Order, withdrawal.Sum, time.Now())
	if err != nil {
		return err
	}

	return nil
}

func (s *DBStorage) WithdrawalHistory(ctx context.Context, user models.User) ([]models.Withdraw, error) {
	var withdrawals []models.Withdraw

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	rows, err := s.DB.QueryContext(ctx, "SELECT number, amount, processed FROM withdrawals WHERE user_id = $1 ORDER BY processed DESC ", user.ID)
	if err != nil {
		return withdrawals, err
	}

	for rows.Next() {
		withdrawal := models.Withdraw{}
		err := rows.Scan(&withdrawal.Order, &withdrawal.Sum, &withdrawal.Time)
		if err != nil {
			return withdrawals, err
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err := rows.Err(); err != nil {
		return withdrawals, err
	}

	return withdrawals, nil
}

func (s *DBStorage) AddOrder(ctx context.Context, order models.Order) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	row := s.DB.QueryRowContext(ctx, `SELECT user_id FROM orders WHERE number = $1 LIMIT 1`, order.Number)

	orderDB := models.Order{}

	err := row.Scan(&orderDB.UserID)
	if err == nil {
		if orderDB.UserID == order.UserID {
			return dbErr.ErrAlreadyLoaded
		} else {
			return dbErr.ErrLoadedByOtherUser
		}
	}

	_, err = s.DB.ExecContext(ctx, `INSERT INTO orders (user_id, number, status, accrual, uploaded) VALUES ($1, $2, $3, $4, $5)`, order.UserID, order.Number, "NEW", 0, time.Now())
	if err != nil {
		return err
	}

	err = s.Queue.PushBackOrders(order)
	if err != nil {
		return err
	}

	return nil
}

func (s *DBStorage) OrdersHistory(ctx context.Context, user models.User) ([]models.Order, error) {
	var orders []models.Order

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	rows, err := s.DB.QueryContext(ctx, "SELECT number, status, accrual, uploaded FROM orders WHERE user_id = $1 ORDER BY uploaded DESC ", user.ID)
	if err != nil {
		return orders, err
	}

	for rows.Next() {
		order := models.Order{}
		err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.EventTime)
		if err != nil {
			return orders, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return orders, err
	}

	return orders, nil
}

func (s *DBStorage) GetOrderForUpdate() (models.Order, error) {
	return s.Queue.GetOrder()
}

func (s *DBStorage) GetOrdersForUpdate(ctx context.Context) ([]models.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	var orders []models.Order

	rows, err := s.DB.QueryContext(ctx, `SELECT user_id, number, status FROM orders WHERE status = 'NEW' OR status = 'PROCESSING' ORDER BY uploaded ASC `)
	if err != nil {
		return orders, err
	}

	for rows.Next() {
		order := models.Order{}
		err = rows.Scan(&order.UserID, &order.Number, &order.Status)
		if err != nil {
			return orders, err
		}
		orders = append(orders, order)
	}

	err = rows.Err()
	if errors.Is(err, sql.ErrNoRows) || err == nil {
		return orders, nil
	}

	return orders, err
}

func (s *DBStorage) UpdateOrders(ctx context.Context, orders ...models.Order) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	tx, err := s.DB.Begin()
	defer tx.Rollback()

	if err != nil {
		return err
	}

	stmtOrders, err := tx.PrepareContext(ctx, `UPDATE orders SET status = $1, accrual = $2 WHERE number = $3`)
	if err != nil {
		return err
	}
	defer stmtOrders.Close()

	stmtUsers, err := tx.PrepareContext(ctx, `UPDATE users SET balance = balance + $1 WHERE id = $2`)
	if err != nil {
		return err
	}
	defer stmtUsers.Close()

	for _, order := range orders {
		if _, err := stmtOrders.ExecContext(ctx, order.Status, order.Accrual, order.Number); err != nil {
			return err
		}

		if order.Accrual > 0 {
			if _, err := stmtUsers.ExecContext(ctx, order.Accrual, order.UserID); err != nil {
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *DBStorage) GetConfig() config.Config {
	return s.Cfg
}

func (s *DBStorage) PushFrontOrders(orders ...models.Order) error {
	return s.Queue.PushFrontOrders(orders...)
}

func (s *DBStorage) PushBackOrders(orders ...models.Order) error {
	return s.Queue.PushBackOrders(orders...)
}
