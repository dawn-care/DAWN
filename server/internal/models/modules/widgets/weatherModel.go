package widgets

import (
	"database/sql"
	"fmt"
)

type WeatherWidget struct {
	ID        int     `json:"id"`
	LATITUDE  float32 `json:"latitude"`
	LONGITUDE float32 `json:"longitude"`
}

var DAYS_OF_WEEK_STR = [7]string{"Sun", "Mon", "Tues", "Wed", "Thur", "Fri", "Sat"}
var MONTH_STR = [12]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

// Add Weather Widget
func CreateWeather(db *sql.DB, weather_widget *WeatherWidget) error {
	query := `
			INSTERT INTO "WeatherWidget" (latitude, longitude)
			VALUES ($1, $2)
			RETURNING id;
	`
	return db.QueryRow(query, weather_widget.LATITUDE, weather_widget.LONGITUDE).Scan(&weather_widget.ID)
}

// Add Multiple Weather Widgets
func CreateWeathersBulk(db *sql.DB, weather_widgets []WeatherWidget) error {
	transaction, err := db.Begin()
	if err != nil {
		return err
	}
	statement, err := transaction.Prepare(`
			INSERT INTO "WeatherWidget" (latitude, longitude)
			VALUES ($1, $2)
			RETURNING id;
	`)
	if err != nil {
		return err
	}
	defer statement.Close()

	for i := range weather_widgets {
		err := statement.QueryRow(weather_widgets[i].LATITUDE, weather_widgets[i].LONGITUDE).Scan(&weather_widgets[i].ID)
		if err != nil {
			transaction.Rollback()
			return err
		}
	}
	return transaction.Commit()
}

// Get One Weather Widget by ID
func GetWeatherByID(db *sql.DB, id int) (*WeatherWidget, error) {
	weather_widget := &WeatherWidget{}

	query := `
			SELECT id, latitude, longitude
			FROM "WeatherWidget"
			WHERE id = $1;
	`
	err := db.QueryRow(query, id).Scan(&weather_widget.ID, &weather_widget.LATITUDE, &weather_widget.LONGITUDE)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return weather_widget, nil
}

// Get Multiple Weather Widgets by Day and User
func GetWeatherByDayAndUser(db *sql.DB, carepartner_id int, day, month, day_of_week int) ([]WeatherWidget, error) {
	dayStr := fmt.Sprintf("%d", day)                        //For Monthly Type - int 1 - 31
	dayOfWeekStr := DAYS_OF_WEEK_STR[day_of_week]           //For Weekly Type - string "Sun" - "Sat"
	monthDay := fmt.Sprintf("%s %d", MONTH_STR[month], day) //For Yearly Type - string "Month day" - i.e. "Jan 1"
	var weather_widgets []WeatherWidget

	query := `
			SELECT widget.id, widget.latitude, widget.longitude
			FROM "Configuration" config
			JOIN "Module" mod ON config.module_id = mod.id
			JOIN "WeatherWidget" widget ON mod.metadata_id = widget.id
			WHERE config.carepartner_id = $1
			AND (
				config.occurence_type = 'Daily'
				OR (config.occurence_type = 'Weekly' AND $2 = ANY(config.occurence_definition))
				OR (config.occurence_type = 'Monthly' AND $3 = ANY(config.occurence_definition))
				OR (config.occurence_type = 'Yearly' AND $4 = ANY(config.occurence_definition))
			)
			AND mod.module_type = 'WeatherWidget'
			ORDER BY config.position;
	`
	rows, err := db.Query(query, carepartner_id, dayOfWeekStr, dayStr, monthDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var weather_widget WeatherWidget
	for rows.Next() {
		err := rows.Scan(&weather_widget.ID, &weather_widget.LATITUDE, &weather_widget.LONGITUDE)
		if err != nil {
			return nil, fmt.Errorf("Error Scanning: %v", err)
		}
		weather_widgets = append(weather_widgets, weather_widget)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error in rows: %v", err)
	}

	return weather_widgets, nil
}

// Get All Weather Widgets by User
func GetUserWeather(db *sql.DB, carepartner_id int) ([]WeatherWidget, error) {
	var weather_widgets []WeatherWidget
	query := `
			SELECT widget.id, widget.latitude, widget.longitude
			FROM "Module" AS mod
			JOIN "WeatherWidget" AS widget ON mod.metadat_id = widget.id
			WHERE mod.carepartner_id = $1
				AND mod.module_type = 'WeatherWidget';
	`
	rows, err := db.Query(query, carepartner_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var weather_widget WeatherWidget
	for rows.Next() {
		err := rows.Scan(&weather_widget.ID, &weather_widget.LATITUDE, &weather_widget.LONGITUDE)
		if err != nil {
			return nil, fmt.Errorf("Error Scanning: %v", err)
		}
		weather_widgets = append(weather_widgets, weather_widget)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Rows Error: %v", err)
	}

	return weather_widgets, nil
}

// Update a Weather Widget
func UpdateWeather(db *sql.DB, weather_widget *WeatherWidget) error {
	query := `
			UPDATE "WeatherWidget"
			SET latitude = $1, longitude = $2
			WHERE id = $4;
	`
	result, err := db.Exec(query, weather_widget.LATITUDE, weather_widget.LONGITUDE, weather_widget.ID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if count == 0 {
		return fmt.Errorf("No Weather Widget Found With ID: %d", weather_widget.ID)
	}
	return nil
}

// Delete a Weather Widget
func DeleteWeather(db *sql.DB, id int) error {
	query := `
			DELETE FROM "WeatherWidget"
			WHERE id = $1;
	`
	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if count == 0 {
		return fmt.Errorf("No Weather Widget Found With ID: %d", id)
	}
	return nil
}

// Get Weather Widget Stats (i.e Total Weather Widgets)
func GetWeatherStats(db *sql.DB) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM "WeatherWidget";`
	err := db.QueryRow(query).Scan(&count)
	return count, err
}
