package widgets

import (
	"database/sql"
	"fmt"
)

/*
WeatherWidget contains information that is required by the weather widget
ID - the ID of the weather widget in the database
ZIPCODE - the zipcode of the carepartner so their weather information is accurate to their location
*/
type WeatherWidget struct {
	ID      int    `json:"id"`
	ZIPCODE string `json:"zipcode"`
}

var DAYS_OF_WEEK_STR = [7]string{"Sun", "Mon", "Tues", "Wed", "Thur", "Fri", "Sat"}
var MONTH_STR = [12]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

// CREATE OPERATIONS

/*
Add a new weather widget
db - instance of the database
weather_widget - the weather widge to be added
*/
func CreateWeather(db *sql.DB, weather_widget *WeatherWidget) error {
	query := `
			INSERT INTO "WeatherWidget" (zipcode)
			VALUES ($1)
			RETURNING id;
	`
	return db.QueryRow(query, weather_widget.ZIPCODE).Scan(&weather_widget.ID)
}

/*
Add multiple weather widgets at once (ex. for their location and their caregivers location, their family, etc.)
db - instance of the database
weather_widgets - array of weather widgets to be added
*/
func CreateWeathersBulk(db *sql.DB, weather_widgets []WeatherWidget) error {
	transaction, err := db.Begin()
	if err != nil {
		return err
	}
	statement, err := transaction.Prepare(`
			INSERT INTO "WeatherWidget" (zipcode)
			VALUES ($1)
			RETURNING id;
	`)
	if err != nil {
		return err
	}
	defer statement.Close()

	for i := range weather_widgets {
		err := statement.QueryRow(weather_widgets[i].ZIPCODE).Scan(&weather_widgets[i].ID)
		if err != nil {
			transaction.Rollback()
			return err
		}
	}
	return transaction.Commit()
}

// RETRIEVE OPERATIONS

/*
Get one weather widget by the ID in the database
db - instance of the database
id - the ID of the weather widget entry
*/
func GetWeatherByID(db *sql.DB, id int) (*WeatherWidget, error) {
	weather_widget := &WeatherWidget{}

	query := `
			SELECT id, zipcode
			FROM "WeatherWidget"
			WHERE id = $1;
	`
	err := db.QueryRow(query, id).Scan(&weather_widget.ID, &weather_widget.ZIPCODE)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return weather_widget, nil
}

// Retrieve all Weather Widgets for a given carepartner and caregiver
func GetWeatherByCarepartnerAndCaregiver(db *sql.DB, carepartner_id int, caregiver_id int) ([]WeatherWidget, error) {
	query := `
			SELECT weather.id, weather.zipcode
			FROM "WeatherWidget" weather
			JOIN "Module" mod ON mod.metadata_id = weather.id
			WHERE mod.carepartner_id = $1 AND mod.caregiver_id = $2;
	`
	rows, err := db.Query(query, carepartner_id, caregiver_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var weather_widgets []WeatherWidget
	for rows.Next() {
		var weather_widget WeatherWidget
		if err := rows.Scan(&weather_widget.ID, &weather_widget.ZIPCODE); err != nil {
			return nil, err
		}
		weather_widgets = append(weather_widgets, weather_widget)
	}
	return weather_widgets, nil
}

// Get Multiple Weather Widgets by Day and User
func GetWeatherByDayAndCarepartner(db *sql.DB, carepartner_id int, day, month, day_of_week int) ([]WeatherWidget, error) {
	dayStr := fmt.Sprintf("%d", day)                        //For Monthly Type - int 1 - 31
	dayOfWeekStr := DAYS_OF_WEEK_STR[day_of_week]           //For Weekly Type - string "Sun" - "Sat"
	monthDay := fmt.Sprintf("%s %d", MONTH_STR[month], day) //For Yearly Type - string "Month day" - i.e. "Jan 1"
	var weather_widgets []WeatherWidget

	query := `
			SELECT widget.id, widget.zipcode
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
		err := rows.Scan(&weather_widget.ID, &weather_widget.ZIPCODE)
		if err != nil {
			return nil, fmt.Errorf("error scanning: %v", err)
		}
		weather_widgets = append(weather_widgets, weather_widget)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error in rows: %v", err)
	}

	return weather_widgets, nil
}

/*
Get all weather widgets for a specific user
db - instance of the database
carepartner_id - the ID of the carepartner who the weather widgets are for
*/
func GetUserWeather(db *sql.DB, carepartner_id int) ([]WeatherWidget, error) {
	var weather_widgets []WeatherWidget
	query := `
			SELECT widget.id, widget.zipcode
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
		err := rows.Scan(&weather_widget.ID, &weather_widget.ZIPCODE)
		if err != nil {
			return nil, fmt.Errorf("error scanning: %v", err)
		}
		weather_widgets = append(weather_widgets, weather_widget)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return weather_widgets, nil
}

// UPDATE OPERATION

/*
Update the weather widget
db - instance of the database
weather_widget - the weather widget that should be updated
*/
func UpdateWeather(db *sql.DB, weather_widget *WeatherWidget) error {
	query := `
			UPDATE "WeatherWidget"
			SET zipcode = $1
			WHERE id = $2;
	`
	result, err := db.Exec(query, weather_widget.ZIPCODE, weather_widget.ID)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no weather widget found with id: %d", weather_widget.ID)
	}
	return nil
}

// DELETE OPERATION

/*
Delete a weather widget
db - instance of the database
id - the ID of the weather widget to be deleted
*/
func DeleteWeather(db *sql.DB, id int) error {
	query := `
			DELETE FROM "WeatherWidget"
			WHERE id = $1;
	`
	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return fmt.Errorf("no weather widget found with id: %d", id)
	}
	return nil
}

// ANALYTICS & STATS

/*
Get weather widget stats (total count of weather widgets)
db - instance of the database
*/
func GetWeatherStats(db *sql.DB) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM "WeatherWidget";`
	err := db.QueryRow(query).Scan(&count)
	return count, err
}
