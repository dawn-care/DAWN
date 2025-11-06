package main

import (
	"fmt"
	"log"
	"server/internal/models/modules/prompts"
	"testing"
	"time"
)

// Uses the same ConnectToDB as caregiver_location_model_test

// Insert an new event list
func TestCreateEventList(t *testing.T) {
	db := ConnectToDB(t)
	defer db.Close()

	eventList := []*prompts.EventPrompt{
		{List_ID: 123456789, Time: time.Now(), Text: "Take Meds"},
		{List_ID: 123456789, Time: time.Now(), Text: "Grandkids coming to visit"},
		{List_ID: 123456789, Time: time.Now(), Text: "Lunch"},
	}
	if err := prompts.CreateEventList(db, eventList); err != nil {
		log.Fatal("CreateEventList failed:", err)
	}
	fmt.Println("Inserted Event List:", eventList[0].List_ID)

	// Clean up
	if err := prompts.DeleteEventList(db, &eventList[0].List_ID); err != nil {
		log.Fatal("Cleanup failed:", err)
	}
}

// Get Event List by list_id
func TestGetEventList(t *testing.T) {
	db := ConnectToDB(t)
	defer db.Close()

	eventList := []*prompts.EventPrompt{
		{List_ID: 123456789, Time: time.Now(), Text: "Take Meds"},
		{List_ID: 123456789, Time: time.Now(), Text: "Grandkids coming to visit"},
		{List_ID: 123456789, Time: time.Now(), Text: "Lunch"},
	}
	if err := prompts.CreateEventList(db, eventList); err != nil {
		log.Fatal("CreateEventList failed:", err)
	}

	got, err := prompts.GetEventListByListID(db, &eventList[0].List_ID)
	if err != nil {
		log.Fatal("GetEventListByListID failed:", err)
	}
	fmt.Println("Got List:", got)

	// Clean up
	if err := prompts.DeleteEventList(db, &eventList[0].List_ID); err != nil {
		log.Fatal("Cleanup failed:", err)
	}
}

// Update Event
func TestUpdateEvent(t *testing.T) {
	db := ConnectToDB(t)
	defer db.Close()

	eventList := []*prompts.EventPrompt{
		{List_ID: 123456789, Time: time.Now(), Text: "Take Meds"},
		{List_ID: 123456789, Time: time.Now(), Text: "Grandkids coming to visit"},
		{List_ID: 123456789, Time: time.Now(), Text: "Lunch"},
	}
	if err := prompts.CreateEventList(db, eventList); err != nil {
		log.Fatal("CreateEventList failed:", err)
	}

	got, err := prompts.GetEventListByListID(db, &eventList[0].List_ID)
	if err != nil {
		log.Fatal("GetEventListByListID failed:", err)
	}
	var newEvent = &got[0]
	newEvent.Text = "Don't take meds"

	if err := prompts.UpdateEvent(db, newEvent); err != nil {
		log.Fatal("UpdateEvent failed:", err)
	}
	fmt.Println("Updated event with id:", newEvent.ID)
	// Clean up
	if err := prompts.DeleteEventList(db, &eventList[0].List_ID); err != nil {
		log.Fatal("Cleanup failed:", err)
	}
}

// Delete Event List
func TestDeleteEventList(t *testing.T) {
	db := ConnectToDB(t)
	defer db.Close()

	eventList := []*prompts.EventPrompt{
		{List_ID: 123456789, Time: time.Now(), Text: "Take Meds"},
		{List_ID: 123456789, Time: time.Now(), Text: "Grandkids coming to visit"},
		{List_ID: 123456789, Time: time.Now(), Text: "Lunch"},
	}
	if err := prompts.CreateEventList(db, eventList); err != nil {
		log.Fatal("CreateEventList failed:", err)
	}

	if err := prompts.DeleteEventList(db, &eventList[0].List_ID); err != nil {
		log.Fatal("DeleteEventList failed:", err)
	}
	fmt.Println("Deleted Event List:", eventList[0].List_ID)
}
