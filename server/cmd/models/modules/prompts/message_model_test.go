package main

import (
	"fmt"
	"log"
	"server/internal/models/modules/prompts"
	"testing"
)

// Uses the same ConnectToDB as caregiver_location_model_test

// Insert single message
func TestCreateMessage(t *testing.T) {
	db := ConnectToDB(t)
	defer db.Close()

	message := &prompts.MessagePrompt{
		Message: "We love you!!",
		URI:     "C:\\image1.png",
	}
	if err := prompts.CreateMessage(db, message); err != nil {
		log.Fatal("CreateMessage failed:", err)
	}
	fmt.Println("Inserted message:", message.ID)

	// Clean up
	if err := prompts.DeleteMessage(db, message.ID); err != nil {
		log.Fatal("Cleanup failed:", err)
	}
}

// Read message by ID
func TestReadMessageByID(t *testing.T) {
	db := ConnectToDB(t)
	defer db.Close()

	message := &prompts.MessagePrompt{
		Message: "We love you!!",
		URI:     "C:\\image1.png",
	}
	if err := prompts.CreateMessage(db, message); err != nil {
		log.Fatal("CreateMessage failed:", err)
	}

	got, err := prompts.GetMessageByID(db, message.ID)
	if err != nil {
		log.Fatal("GetMessageByID failed:", err)
	}
	fmt.Println("Got message:", got)

	// Clean up
	if err := prompts.DeleteMessage(db, message.ID); err != nil {
		log.Fatal("Cleanup failed:", err)
	}
}

// Update message
func TestUpdateMessage(t *testing.T) {
	db := ConnectToDB(t)
	defer db.Close()

	message := &prompts.MessagePrompt{
		Message: "We love you!!",
		URI:     "C:\\image1.png",
	}
	if err := prompts.CreateMessage(db, message); err != nil {
		log.Fatal("CreateMessage failed:", err)
	}

	message.Message = "Don't forget to drink water!"
	if err := prompts.UpdateMessagePrompt(db, message); err != nil {
		log.Fatal("UpdateMessagePrompt failed:", err)
	}
	fmt.Println("Updated message:", message.ID)

	// Clean up
	if err := prompts.DeleteMessage(db, message.ID); err != nil {
		log.Fatal("Cleanup failed:", err)
	}
}

// Delete message
func TestDeleteMessage(t *testing.T) {
	db := ConnectToDB(t)
	defer db.Close()

	message := &prompts.MessagePrompt{
		Message: "We love you!!",
		URI:     "C:\\image1.png",
	}
	if err := prompts.CreateMessage(db, message); err != nil {
		log.Fatal("CreateMessage failed:", err)
	}

	if err := prompts.DeleteMessage(db, message.ID); err != nil {
		log.Fatal("DeleteMessage failed:", err)
	}
	fmt.Println("Deleted message:", message.ID)
}
