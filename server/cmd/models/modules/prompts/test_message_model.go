package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"server/internal/models/modules/prompts"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func TestMessageModel() {
	godotenv.Load("../../../../.env")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Insert single message
	message := &prompts.MessagePrompt{
		Message: "We love you!!",
		URI:     "C:\\image1.png",
	}
	if err := prompts.CreateMessage(db, message); err != nil {
		log.Fatal("CreateMessage failed:", err)
	}
	fmt.Println("Inserted message:", message.ID)

	// Read message by ID
	got, err := prompts.GetMessageByID(db, message.ID)
	if err != nil {
		log.Fatal("GetMessageByID failed:", err)
	}
	fmt.Println("Got message:", got)

	// Update message
	message.Message = "Don't forget to drink water!"
	if err := prompts.UpdateMessagePrompt(db, message); err != nil {
		log.Fatal("UpdateMessagePrompt failed:", err)
	}
	fmt.Println("Updated message:", message.ID)

	// Delete message
	if err := prompts.DeleteMessage(db, message.ID); err != nil {
		log.Fatal("DeleteMessage failed:", err)
	}
	fmt.Println("Deleted message:", message.ID)
}
