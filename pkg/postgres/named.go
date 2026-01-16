package postgres

import (
	"context"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var namedRegexp = regexp.MustCompile(`@(\w+)`)

func QueryNamed(ctx context.Context, pool *pgxpool.Pool, query string, args map[string]interface{}) (pgx.Rows, error) {
	query, positionalArgs := convertNamedQuery(query, args)

	return pool.Query(ctx, query, positionalArgs...)
}

func QueryRowNamed(ctx context.Context, pool *pgxpool.Pool, query string, args map[string]interface{}) pgx.Row {
	query, positionalArgs := convertNamedQuery(query, args)
	return pool.QueryRow(ctx, query, positionalArgs...)
}

func ExecNamed(ctx context.Context, pool *pgxpool.Pool, query string, args map[string]interface{}) (pgconn.CommandTag, error) {
	query, positionalArgs := convertNamedQuery(query, args)
	return pool.Exec(ctx, query, positionalArgs...)
}

func convertNamedQuery(query string, args map[string]interface{}) (string, []interface{}) {
	var positionalArgs []interface{}
	argIndex := make(map[string]int)

	convertedQuery := namedRegexp.ReplaceAllStringFunc(query, func(match string) string {
		paramName := match[1:] // убираем "@"

		if idx, ok := argIndex[paramName]; ok {
			return fmt.Sprintf("$%d", idx)
		}

		positionalArgs = append(positionalArgs, args[paramName])
		argIndex[paramName] = len(positionalArgs)
		return fmt.Sprintf("$%d", len(positionalArgs))
	})

	return convertedQuery, positionalArgs
}
