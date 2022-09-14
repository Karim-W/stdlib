package stdlib

import (
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "secret"
	dbname   = "temp"
)

// func TestNativeDbCon(t *testing.T) {
// 	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
// 		"password=%s dbname=%s sslmode=disable",
// 		host, port, user, password, dbname)
// 	d := NativeDatabaseProvider("postgres", psqlInfo)
// 	res, err := d.ExecContext(NewContext(), "DROP TABLE IF EXISTS tt")
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(res)
// 	res, err = d.ExecContext(NewContext(), "INSERT INTO vibes VALUES (3,'hello',1,'ABC')")
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(res)
// 	res, err = d.ExecContext(NewContext(), "INSERT INTO vibes VALUES (4,'hello',2,'ABC')")
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(res)

// }
