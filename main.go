package main

import (
    "encoding/json"
    "context"
	"github.com/gorilla/mux"
    "go.elastic.co/apm/module/apmgorilla"
    "log"
    "os"
	"net/http"
    "github.com/go-redis/redis"
    "fmt"
    "go.elastic.co/apm/module/apmgoredis"
    "go.elastic.co/apm/module/apmsql"
    _ "go.elastic.co/apm/module/apmsql"
    _ "go.elastic.co/apm/module/apmsql/pq"
    _ "github.com/lib/pq"
    _ "github.com/kr/pretty"
    "go.elastic.co/apm"
)

type Author struct {
	Name string `json:"name"`
	Age int     `json:"age"`
}


const (
  host     = os.Getenv("host")
  port     = os.Getenv("port")
  user     = os.Getenv("user")
  password = os.Getenv("password")
  dbname   = os.Getenv("dbname")
)

func main() {
	r := mux.NewRouter()
    apmgorilla.Instrument(r)
	r.HandleFunc("/hello", funcHandler)
    r.HandleFunc("/redis", funcRedis)
    r.HandleFunc("/db", funcPostgres)

	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatal("Unable to start service")
	}
}

func funcHandler(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Ok",
	})
}

func funcRedis(w http.ResponseWriter, r *http.Request){
    client := apmgoredis.Wrap(redis.NewClient(&redis.Options{
        Addr: os.Getenv("redis"),
        Password: "",
		DB: 0,
    })).WithContext(r.Context())
   
    json, err := json.Marshal(Author{Name: "Elliot", Age: 25})
    if err != nil {
        fmt.Println(err)
    }

    err = client.Set("id1234", json, 0).Err()
    if err != nil {
        fmt.Println(err)
    }
    val, err := client.Get("id1234").Result()
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println(val)
}


func funcPostgres(w http.ResponseWriter, r *http.Request){
    ctx := r.Context()
	processingRequest(ctx)    
}


func processingRequest(ctx context.Context) {
	span, ctx := apm.StartSpan(ctx, "getListOrders", "custom")
	defer span.End()

    psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",host, port, user, password, dbname)
    db, err := apmsql.Open("postgres", psqlInfo)
    if err != nil {
        panic(err)
    }
    defer db.Close()

    fmt.Println("Successfully connected!")
    
    rows, err := db.QueryContext(ctx,"select order_id,order_date from public.order")
    if err != nil {
		log.Fatal(err)
    }

    cs, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
    }
    fmt.Printf("%v\n", cs)
	return
}