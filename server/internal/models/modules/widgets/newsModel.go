package widgets

import (
	"database/sql"
	"fmt"
)

type NewsWidget struct {
	ID       int    `json:"id"`
	CATEGORY string `json:"category"`
	LANGUAGE string `json:"language"`
	COUNTRY  string `json:"country"`
}

// Add News Widget
func CreateNews(db *sql.DB, news_widget *NewsWidget) error {
	query := `
		INSERT INTO "NewsWidget" (category, language, country)
		VALUES ($1, $2, $3)
		RETURNING id;
	`
	return db.QueryRow(query, news_widget.CATEGORY, news_widget.LANGUAGE, news_widget.COUNTRY).Scan(&news_widget.ID)
}

// Add Multiple News Widgets
func CreateNewsBulk(db *sql.DB, news_widgets []NewsWidget) error {
	transaction, err := db.Begin()
	if err != nil {
		return err
	}
	statement, err := transaction.Prepare(`
		INSERT INTO "NewsWidget" (category, language, country)
		VALUES ($1, $2, $3)
		RETURNING id;
	`)
	if err != nil {
		return err
	}
	defer statement.Close()

	for i := range news_widgets {
		err := statement.QueryRow(
			news_widgets[i].CATEGORY,
			news_widgets[i].LANGUAGE,
			news_widgets[i].COUNTRY,
		).Scan(&news_widgets[i].ID)
		if err != nil {
			transaction.Rollback()
			return err
		}
	}
	return transaction.Commit()
}

// Get One News Widget by ID
func GetNewsByID(db *sql.DB, id int) (*NewsWidget, error) {
	news_widget := &NewsWidget{}
	query := `
		SELECT id, category, language, country
		FROM "NewsWidget"
		WHERE id = $1;
	`
	err := db.QueryRow(query, id).Scan(&news_widget.ID, &news_widget.CATEGORY, &news_widget.LANGUAGE, &news_widget.COUNTRY)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return news_widget, nil
}

// Get Multiple News Widgets by Day and User
func GetNewsByDayAndCarepartner(db *sql.DB, carepartner_id int, day, month, day_of_week int) ([]NewsWidget, error) {
	dayStr := fmt.Sprintf("%d", day)                        // For Monthly
	dayOfWeekStr := DAYS_OF_WEEK_STR[day_of_week]           // For Weekly
	monthDay := fmt.Sprintf("%s %d", MONTH_STR[month], day) // For Yearly
	var news_widgets []NewsWidget

	query := `
		SELECT widget.id, widget.category, widget.language, widget.country
		FROM "Configuration" AS config
		JOIN "Module" AS mod ON config.module_id = mod.id
		JOIN "NewsWidget" AS widget ON mod.metadata_id = widget.id
		WHERE config.carepartner_id = $1
		AND (
			config.occurence_type = 'Daily'
			OR (config.occurence_type = 'Weekly' AND $2 = ANY(config.occurence_definition))
			OR (config.occurence_type = 'Monthly' AND $3 = ANY(config.occurence_definition))
			OR (config.occurence_type = 'Yearly' AND $4 = ANY(config.occurence_definition))
		)
		AND mod.module_type = 'NewsWidget'
		ORDER BY config.position;
	`
	rows, err := db.Query(query, carepartner_id, dayOfWeekStr, dayStr, monthDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var news_widget NewsWidget
	for rows.Next() {
		err := rows.Scan(&news_widget.ID, &news_widget.CATEGORY, &news_widget.LANGUAGE, &news_widget.COUNTRY)
		if err != nil {
			return nil, fmt.Errorf("Error Scanning: %v", err)
		}
		news_widgets = append(news_widgets, news_widget)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("Error in rows: %v", err)
	}

	return news_widgets, nil
}

// Get All News Widgets by User
func GetUserNews(db *sql.DB, carepartner_id int) ([]NewsWidget, error) {
	var news_widgets []NewsWidget
	query := `
		SELECT widget.id, widget.category, widget.language, widget.country
		FROM "Module" AS mod
		JOIN "NewsWidget" AS widget ON mod.metadata_id = widget.id
		WHERE mod.carepartner_id = $1
			AND mod.module_type = 'NewsWidget';
	`
	rows, err := db.Query(query, carepartner_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var news_widget NewsWidget
	for rows.Next() {
		err := rows.Scan(&news_widget.ID, &news_widget.CATEGORY, &news_widget.LANGUAGE, &news_widget.COUNTRY)
		if err != nil {
			return nil, fmt.Errorf("Error Scanning: %v", err)
		}
		news_widgets = append(news_widgets, news_widget)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("Rows Error: %v", err)
	}

	return news_widgets, nil
}

// Update a News Widget
func UpdateNews(db *sql.DB, news_widget *NewsWidget) error {
	query := `
		UPDATE "NewsWidget"
		SET category = $1, language = $2, country = $3
		WHERE id = $4;
	`
	result, err := db.Exec(query, news_widget.CATEGORY, news_widget.LANGUAGE, news_widget.COUNTRY, news_widget.ID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if count == 0 {
		return fmt.Errorf("No News Widget Found With ID: %d", news_widget.ID)
	}
	return nil
}

// Delete a News Widget
func DeleteNews(db *sql.DB, id int) error {
	query := `
		DELETE FROM "NewsWidget"
		WHERE id = $1;
	`
	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if count == 0 {
		return fmt.Errorf("No News Widget Found With ID: %d", id)
	}
	return nil
}

// Get News Widget Stats (i.e., total count)
func GetNewsStats(db *sql.DB) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM "NewsWidget";`
	err := db.QueryRow(query).Scan(&count)
	return count, err
}
