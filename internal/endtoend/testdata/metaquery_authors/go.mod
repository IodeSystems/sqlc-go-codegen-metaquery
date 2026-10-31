module github.com/iodesystems/sqlc-go-codegen-metaquery/internal/endtoend/testdata/metaquery_authors

go 1.26.0

require (
	github.com/iodesystems/sqlc-go-codegen-metaquery v1.7.0
	github.com/jackc/pgx/v5 v5.12.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	golang.org/x/text v0.41.0 // indirect
)

replace github.com/iodesystems/sqlc-go-codegen-metaquery => ../../../..
