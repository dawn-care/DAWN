package widgets

import (
	"testing"

	_ "github.com/lib/pq"
	// "github.com/joho/godotenv"
)

// Helper Function to Clear Test News Table
func clearNewsTable(t *testing.T) {
	_, err := testDB.Exec(`DELETE FROM "NewsWidget";`)
	if err != nil {
		t.Fatalf("Failed to clear table: %v", err)
	}
}

// Tests CreateNews and GetNewsByID
func TestCreateNewsWidget(t *testing.T) {
	clearNewsTable(t)

	news_widget := &NewsWidget{
		CATEGORY: "Animals",
		LANGUAGE: "en",
		COUNTRY:  "us",
	}

	err := CreateNews(testDB, news_widget)
	if err != nil {
		t.Fatalf("CreateNews Failed: %v", err)
	}
	if news_widget.ID == 0 {
		t.Errorf("News Widget ID Not Set")
	}

	found_widget, err := GetNewsByID(testDB, news_widget.ID)
	if err != nil {
		t.Fatalf("GetNewsByID Failed: %v", err)
	}
	if found_widget == nil {
		t.Fatalf("Widget Not Found")
	}
	if found_widget.CATEGORY != "Animals" {
		t.Errorf("Unexpected Widget Returned (CreateNews): %+v", found_widget)
	}
	if found_widget.LANGUAGE != "en" {
		t.Errorf("Unexpected Widget Returned (CreateNews): %+v", found_widget)
	}
	if found_widget.COUNTRY != "us" {
		t.Errorf("Unexpected Widget Returned (CreateNews): %+v", found_widget)
	}
}

// Tests UpdateNews
func TestUpdateNewsWidget(t *testing.T) {
	clearNewsTable(t)

	news_widget := &NewsWidget{
		CATEGORY: "Animals",
		LANGUAGE: "en",
		COUNTRY:  "us",
	}

	err := CreateNews(testDB, news_widget)
	if err != nil {
		t.Fatalf("CreateNews Failed: %v", err)
	}

	news_widget.LANGUAGE = "es"
	err = UpdateNews(testDB, news_widget)
	if err != nil {
		t.Fatalf("UpdateNews Failed: %v", err)
	}
	found_widget, _ := GetNewsByID(testDB, news_widget.ID)
	if found_widget.LANGUAGE != "es" {
		t.Errorf("Unexpected Widget Returned (UpdateNews): %+v", found_widget)
	}
}

// Tests DeleteNews

func TestDeleteNewsWidget(t *testing.T) {
	clearNewsTable(t)

	news_widget := &NewsWidget{
		CATEGORY: "Animals",
		LANGUAGE: "en",
		COUNTRY:  "us",
	}

	err := CreateNews(testDB, news_widget)
	if err != nil {
		t.Fatalf("CreateNews Failed: %v", err)
	}

	err = DeleteNews(testDB, news_widget.ID)
	if err != nil {
		t.Fatalf("DeleteNews Failed: %v", err)
	}
	found_widget, _ := GetNewsByID(testDB, news_widget.ID)
	if found_widget != nil {
		t.Errorf("Unexpected Widget Returned (DeleteNews): %+v", found_widget)
	}
}

// Tests CreateNewsBulk
func TestCreateNewsBulkWidgets(t *testing.T) {
	clearNewsTable(t)

	news_widgets := []NewsWidget{
		{CATEGORY: "Animals", LANGUAGE: "en", COUNTRY: "us"},
		{CATEGORY: "Charity", LANGUAGE: "fr", COUNTRY: "ca"},
		{CATEGORY: "Good News", LANGUAGE: "es", COUNTRY: "mx"},
	}
	err := CreateNewsBulk(testDB, news_widgets)
	if err != nil {
		t.Fatalf("CreateNewsBulk Failed: %v", err)
	}
	found_widget, _ := GetNewsByID(testDB, news_widgets[0].ID)
	if found_widget == nil {
		t.Error("CreateNewsBulk widget[0] Not Found")
	}
	found_widget, _ = GetNewsByID(testDB, news_widgets[1].ID)
	if found_widget == nil {
		t.Error("CreateNewsBulk widget[1] Not Found")
	}
	found_widget, _ = GetNewsByID(testDB, news_widgets[2].ID)
	if found_widget == nil {
		t.Error("CreateNewsBulk widget[2] Not Found")
	}
}

// Test GetNewsStats
func TestGetNewsStatsWidget(t *testing.T) {
	clearNewsTable(t)

	news_widgets := []NewsWidget{
		{CATEGORY: "Animals", LANGUAGE: "en", COUNTRY: "us"},
		{CATEGORY: "Charity", LANGUAGE: "fr", COUNTRY: "ca"},
		{CATEGORY: "Good News", LANGUAGE: "es", COUNTRY: "mx"},
	}
	err := CreateNewsBulk(testDB, news_widgets)
	if err != nil {
		t.Fatalf("CreateNewsBulk Failed: %v", err)
	}

	count, _ := GetNewsStats(testDB)
	if count != 3 {
		t.Errorf("Unexpected Count Returned: %d", count)
	}
}
