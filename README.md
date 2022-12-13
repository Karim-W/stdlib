# Stdlib

This is a collection of standard library modules for go i use in my projects. The package offers a set of modules that are useful in most projects such as logging,tracing,pooling,caching,HTTP Request,etc.

## Modules
- [Cache](#cache)
- [HTTP Client](#http-client)
- [SQL](#sql)
- [Pool](#pool)

## Usage

### Importing
in code

```go
import "github.com/karim-w/stdlib"
```
and
```bash
go get github.com/karim-w/stdlib
```

### Cache

Caching tools are provided to make it easier to use caching in your projects. The package provides two types of caching, in memory and redis. The package uses the [go-cache]("github.com/patrickmn/go-cache") package for in memory caching and the [go-redis]("github.com/go-redis/redis") package for redis caching.

#### Redis Cache
uses the [go-redis]("github.com/go-redis/redis") package

```go
cache, err := InitRedisCache("redis://:@localhost:6379/0")
if err != nil {
	t.Error(err)
}
err = cache.SetCtx(context.TODO(),"key",map[string]interface{}{
	"value":8,
})
```
#### In Memory Cache
uses the [go-cache]("github.com/patrickmn/go-cache") package

```go
cache, err := InitInMemoryCache(
		time.Minute,
		time.Minute*2,
	)
if err != nil {
	t.Error(err)
}
err = cache.SetCtx(context.TODO(),"key",map[string]interface{}{
	"value":8,
})
```
#### Extentedable via
##### Logger
use the ZapLogger or any other logger that implements the Logger interface
```go
var logger *zap.Logger
newCache := cache.WithLogger(logger)
```
##### Tracer
currently only support the applicationinsights tracer loacted at [appinsights]("github.com/BetaLixT/appInsightsTrace")
```go
var tracer *tracer.AppInsightsCore
newCache := cache.WithTracer(tracer)
```

### HTTP Client
The package provides a wrapper for the [http]("net/http") package to make it easier to use. The package provides a client that can be used to make HTTP requests. The package also provides a middleware that can be used to add tracing and logging to the requests.

#### Initializing
##### Plain client
```go
var logger *zap.Logger
client := stdlib.ClientProvider()
```
> Note: This function **WILL** panic if it fails to initialize a logger

##### Traceable client
```go
var logger *zap.Logger
var tracer *tracer.AppInsightsCore
client := stdlib.TracedClientProvider(
	tracer,
	logger,
)
```
##### Client Options
```go
type ClientOptions struct {
	Authorization string             //Authorization header
	ContentType   string             //Content-Type header
	Query         string             //Query string to append to the url
	Headers       *map[string]string //Headers to add to the request
	Timeout       *time.Time         //Timeout for the request
}
```

#### Usage
The client provides 5 Main Methods to make requests which are `GET` , `POST` , `PUT` , `DELETE` , `PATCH`. Each method takes a context, url, options, body and response. The response MUST be a pointer to the response type.
`GET` And `Del` do not take a body in the request.
##### GET
Make a GET request
```go
Response := interface{}		//any type
code,err := client.Get(
	context.TODO(),
	"http://localhost:8080",
	&stdlib.ClientOptions{},
	&Response,
)
```
> Note: MUST pass a pointer to the response
##### POST
Make a POST request
```go
Body := interface{}			//any type
Response := interface{}		//any type
code,err := client.Post(
	context.TODO(),
	"http://localhost:8080",
	&stdlib.ClientOptions{},
	Body,
	&Response,
)
```
> Note: MUST pass a pointer to the response

##### PUT
Make a PUT request
```go
Body := interface{}			//any type
Response := interface{}		//any type
code,err := client.Put(
	context.TODO(),
	"http://localhost:8080",
	&stdlib.ClientOptions{},
	Body,
	&Response,
)
```
> Note: MUST pass a pointer to the response
##### PATCH
Make a PATCH request
```go
Body := interface{}			//any type
Response := interface{}		//any type
code,err := client.Patch(
	context.TODO(),
	"http://localhost:8080",
	&stdlib.ClientOptions{},
	Body,
	&Response,
)
```
> Note: MUST pass a pointer to the response

##### DELETE
Make a DELETE request
```go
Response := interface{}		//any type
code,err := client.Del(
	context.TODO(),
	"http://localhost:8080",
	&stdlib.ClientOptions{},
	&Response,
)
```
> Note: MUST pass a pointer to the response

#### Extentedable via
##### Authorization
```go
client.SetAuthHandler(authProvider)
```
Auth provider must implement the `AuthProvider` interface
```go
type AuthProvider interface {
	GetAuthHeader() string
}
```
### SQL
The package provides a wrapper for the [sql]("database/sql") package to make it easier to use. The package provides a client that can be used to make SQL requests. The package also provides a middleware that can be used to add tracing and logging to the requests.

#### Initializing

> Only supports Postgres, planning to add more

> Will not support MSSQL
```go
db := stdlib.NativeDatabaseProvider(
	databaseDriver,
	databaseConnectionString,
)
```
for tracing and logging use the `TracedNativeDBWrapper` function
```go
db := stdlib.TracedNativeDBWrapper(
	databaseDriver,
	databaseConnectionString,
	tracer,
	dbName,
)
dbwithlogger := db.WithLogger(logger)
```
> WARNING: Both functions will panic if they fail to initialize a logger

#### Usage
The client provides 3 Main Methods to make requests which are `Query` , `Exec` , `QueryRow`. Each with a variant for context they all take in query, args.

##### QueryRow
Make a Query request that returns a single row
```go
var res int
row := db.Query(
	"SELECT COUNT(*) FROM table",
	[]interface{}{}...,
)
err:=row.Scan(&res)
```
##### QueryRowContext
Make a Query request with context that returns a single row
```go
var res int
row := db.QueryContext(
	context.TODO(),
	"SELECT COUNT(*) FROM table",
	[]interface{}{}...,
)
err:=row.Scan(&res)
```

##### Exec
Execute a query that does not return a result rather the number of rows affected
```go
res,err := db.Exec(
	"INSERT INTO table VALUES($1,$2)",
	[]interface{}{}...,
)
```
##### ExecContext
Execute a query that does not return a result rather the number of rows affected with context
```go
res,err := db.ExecContext(
	context.TODO(),
	"INSERT INTO table VALUES($1,$2)",
	[]interface{}{}...,
)
```

##### QueryRow
Make a Query request that returns multiple rows
```go
res := []int{}
rows,err := db.Query(
	"SELECT COUNT(*) FROM table",
	[]interface{}{}...,
)
for rows.Next(){
	var i int
	err:=rows.Scan(&i)
	if err!=nil{
		return err
	}
	res = append(res,i)
}
```
##### QueryRowContext
Make a Query request with context that returns multiple rows
```go
res := []int{}
rows,err := db.QueryContext(
	context.TODO(),
	"SELECT COUNT(*) FROM table",
	[]interface{}{}...,
)
for rows.Next(){
	var i int
	err:=rows.Scan(&i)
	if err!=nil{
		return err
	}
	res = append(res,i)
}
```
#### Begin
Begin starts and returns a new transaction Object
```go
tx,err := db.Begin()
```
#### BeginTx
Begin starts and returns a new transaction Object
```go
tx,err := db.BeginTx(context.TODO(),opts) //opts is a *sql.TxOptions
```

#### Extentedable via
##### Logger
```go
newDb := db.withLogger(logger) //*zap.Logger
```

### Pooler
The package provides Pooling utility for any type of object. The package rotates the objects in the pool and provides a way to get the object from the pool in a round robin thread safe manner.

#### Initializing
```go
type customType struct{
	//some fields
}

pool,err := stdlib.NewPooler(
	func () *customType{
		return &customType{}
	},
	&stdlib.PoolerOptions{
		PoolSize: 10,
	},
)
```

#### Usage
##### Get
Get an object from the pool
```go
obj := pool.Get()
```

##### Size
Get the size of the pool
```go
size := pool.Size()
```

##### Clear
Clear the pool
```go
pool.Clear()
```







