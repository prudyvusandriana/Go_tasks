package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"database/sql"
	"github.com/Nastya-Kruglikova/cool_tasks/src/config"
	"github.com/Nastya-Kruglikova/cool_tasks/src/database"
	"github.com/Nastya-Kruglikova/cool_tasks/src/services"
	"github.com/garyburd/redigo/redis"
	"github.com/urfave/negroni"
)

var (
	DB    *sql.DB
	Cache redis.Conn
)

func main() {
	configFile := flag.String("config", "./config.json", "Configuration file in JSON-format")
	flag.Parse()

	if len(*configFile) > 0 {
		config.FilePath = *configFile
	}

	err := config.Load()
	if err != nil {
		log.Fatalf("error while reading config: %s", err)
	}

	f, err := os.OpenFile(config.Config.LogFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	DB, err = database.SetupPostgres(config.Config.Database)
	if err != nil {
		log.Fatalf("eror while loading postgreSQL: %s:", err)
	}

	Cache, err = database.SetupRedis(config.Config.Database)
	if err != nil {
		log.Fatalf("eror while loading redis: %s:", err)
	}

	defer f.Close()

	log.SetOutput(f)

	// setting up web server middlewares
	middlewareManager := negroni.New()
	middlewareManager.Use(negroni.NewRecovery())
	middlewareManager.UseHandler(services.NewRouter())

	log.Println("Starting HTTP listener...")
	err = http.ListenAndServe(config.Config.ListenURL, middlewareManager)
	if err != nil {
		log.Println(err)
	}
	log.Printf("Stop running application: %s", err)
}
