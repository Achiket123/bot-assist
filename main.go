package main

import (
	"bot-assist/auth"
	"bot-assist/ent"
	"bot-assist/ent/generated"
	"bot-assist/telegram"
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func init() {

	data, err := os.ReadFile("./client_secret.json")
	if err != nil {
		log.Default().Panicln(err)
	}

	err = json.Unmarshal(data, &auth.KeyVal)
	if err != nil {
		log.Default().Panicln(err)
	}
	log.Default().Println("INITIALIZED EVERYTHING")
}

func main() {

	databaseURL := "root:8759@tcp(localhost:3306)/bot_assist?parseTime=True"
	ctx := context.Background()
	var err error
	ent.Client, err = generated.Open("mysql", databaseURL)
	if err != nil {
		log.Fatalf("Failing opening connection: %v", err)
	}
	defer func() {
		err := ent.Client.Close()
		if err != nil {
			log.Default().Panicln(err)
		}
	}()

	if err := ent.Client.Schema.Create(ctx); err != nil {
		log.Fatalf("Failed creating schema resources: %v", err)
	}
	r := gin.Default()
	go telegram.Start(auth.KeyVal["TG_BOT_API_TOKEN"].(string))
	log.Default().Println("Successfully connected to MySQL and migrated schema and starting telegram bot.")
	r.GET("/callback", auth.AuthCallBack)
	err = r.Run(":8080")
	if err != nil {
		log.Default().Panicln(err)
	}
	select {}
}
