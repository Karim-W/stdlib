package stdlib

import (
	"fmt"
	"runtime"
	"testing"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "secret"
	dbname   = "dex2"
)

func TestNativeDbCon(t *testing.T) {
	runtime.GOMAXPROCS(10)
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	d := NativeDatabaseProvider("postgres", psqlInfo)
	res, err := d.ExecContext(NewContext(), "DROP TABLE IF EXISTS tt")
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}
