package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	pgxv5 "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var migrationsFS embed.FS

func Run(dsn string) error {
	// 1. открыть соединение database/sql; "pgx5" — имя драйвера,
	//    которое регистрирует сам пакет migrate при импорте
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open db for migrations: %w", err)
	}
	defer db.Close()

	// 2. источник: все встроенные SQL-файлы, "." — корень встроенной ФС
	sourceDriver, err := iofs.New(migrationsFS, ".")
	if err != nil {
		return fmt.Errorf("load embedded migrations: %w", err)
	}

	// 3. драйвер базы поверх соединения; конфиг пустой —
	//    настройки по умолчанию (таблица schema_migrations и т.д.)
	databaseDriver, err := pgxv5.WithInstance(db, &pgxv5.Config{})
	if err != nil {
		return fmt.Errorf("create db driver: %w", err)
	}

	// 4. склеить источник и драйвер в мигратор
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "pgx5", databaseDriver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	// 5. накатить; ErrNoChange — «всё уже применено», это не провал
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
