package widgets

import (
	"database/sql"
	"fmt"
)

/*
NewsWidget represents info required to retrieve news information for the news widget
ID - the ID of the entry inthe database
CATEGORY - the category of news
LANGUAGE - the language of the news
COUNTRY - the country the news should be for
*/
type NewsWidget struct {
	ID       int    `json:"id"`
	CATEGORY string `json:"category"`
	LANGUAGE string `json:"language"`
	COUNTRY  string `json:"country"`
}

// CREATE OPERATIONS

/*
Create a new NewsWidget
db - instance of the database
news_widget - information regarding the type of news
*/
func CreateNews(db *sql.DB, news_widget *NewsWidget) error {
	query := `
		INSERT INTO "NewsWidget" (category, language, country)
		VALUES ($1, $2, $3)
		RETURNING id;
	`
	return db.QueryRow(query, news_widget.CATEGORY, news_widget.LANGUAGE, news_widget.COUNTRY).Scan(&news_widget.ID)
}

/*
Add multiple news widgets at once
db - instance of the database
news_widgets - array of news widgets to be added
*/
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

// RETRIEVE OPERATIONS

/*
Get a single news widget by the ID
db - instance of the database
id - the ID of the news widget
*/
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

/*
Get multiple news widgets by Day and User
db - instance of the database
carepartner_id - the ID of the carepartner responsible for the news widget
day - day the news widget should be applicable to
month - month the news widget should be applicable to
day_of_week - the day of the week the news widget should be applicable to
*/
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
			return nil, fmt.Errorf("error scanning: %s", err)
		}
		news_widgets = append(news_widgets, news_widget)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error in rows: %s", err)
	}

	return news_widgets, nil
}

/*
Get all news widgets for the user
db - instance of the database
carepartner_id - the ID of the carepartner who is responsible for the news widget
*/
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
			return nil, fmt.Errorf("error scanning: %s", err)
		}
		news_widgets = append(news_widgets, news_widget)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows error: %s", err)
	}

	return news_widgets, nil
}

// UPDATE OPERATION

/*
Update a news widget
db - instance of the database
news_widget - news widget content to be updated
*/
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
	count, _ := result.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no news widget found with ID: %d", news_widget.ID)
	}
	return nil
}

// DELETE OPERATION

/*
Deletes the specified news widget
db - instance of the database
id - the ID of the news widget in the database
*/
func DeleteNews(db *sql.DB, id int) error {
	query := `
		DELETE FROM "NewsWidget"
		WHERE id = $1;
	`
	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no news widget found with id: %d", id)
	}
	return nil
}

// ANALYTICS & STATS

/*
Get stats on news widgets, i.e. the total number of them
db - instance of the database
*/
func GetNewsStats(db *sql.DB) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM "NewsWidget";`
	err := db.QueryRow(query).Scan(&count)
	return count, err
}
