package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
)

const TEST_DATA_PREFIX = "perf-"

func main() {
	fmt.Println("Starting database test...")
	testGoDatabaseSql("postgres://cloud_controller:fjLip8fvl0nV97OpvI7pJhSV4KQsmA@localhost:5524/cloud_controller?sslmode=disable")
	fmt.Println("Finished.")
}

func testGoDatabaseSql(connection string) {
	db, err := sql.Open("pgx", connection)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	//cleanupTable(db, ctx, "routes", "host")
	//cleanupTable(db, ctx, "domains", "name")
	cleanupTable(db, ctx, "organizations", "name")

	//rows, err := db.Query(`SELECT "id", "guid" FROM "users"`)
	//CheckError(err)
	//defer rows.Close()
	//
	//for rows.Next() {
	//	var id int
	//	var guid string
	//
	//	err = rows.Scan(&id, &guid)
	//	CheckError(err)
	//
	//	fmt.Println(id, guid)
	//}
}

func cleanupTable(db *sql.DB, ctx context.Context, tableName string, columnName string) {
	statement := fmt.Sprintf("DELETE FROM %s WHERE %s LIKE '%s' ON DELETE CASCADE", tableName, columnName, TEST_DATA_PREFIX+"%")
	log.Printf("Running statement: %s", statement)
	result, err := db.ExecContext(ctx, statement)
	if err != nil {
		log.Fatal(err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Deleted %d rows from '%s'", rows, tableName)
}

//func CheckError(err error) {
//	if err != nil {
//		panic(err)
//	}
//}