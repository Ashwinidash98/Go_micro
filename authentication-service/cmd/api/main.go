package main

import (
	"authentication/data"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "80"

var counts int64

type Config struct {
	// DB     *sql.DB
	// Models data.Models
	Repo   data.Repository
	Client *http.Client
}

func main() {
	log.Println("Starting authentication service")

	//TODO connect to DB

	//connect to DB
	conn := connectToDB()
	if conn == nil {
		log.Panic("Can't connect to postgres!!")
	}

	//set up config
	app := Config{
		// DB:     conn,
		// Models: data.New(conn),
		Client: &http.Client{},
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres not yet ready...")
			counts++
		} else {
			log.Println("Connectted to postgres")
			return connection
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing of for two seconds")
		time.Sleep(2 * time.Second)
		continue
	}
}

func (app *Config) setupRepo(conn *sql.DB) {
	db := data.NewPostgresRepository(conn)
	app.Repo = db
}
