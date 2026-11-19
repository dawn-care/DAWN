package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	neturl "net/url"
	"time"
)

// OpenWeatherBaseURL is where the OpenWeather API lives; tests point a service elsewhere
const OpenWeatherBaseURL = "https://api.openweathermap.org"

// WeatherService fetches live weather for screen widgets.
type WeatherService struct {
	apiKey  string
	client  *http.Client
	baseURL string
}

// Struct to represent the geocoding response from OpenWeather API
type CoordinateResponse struct {
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
	Name string  `json:"name"`
	Zip  string  `json:"zip"`
}

// Struct to represent the overview response from OpenWeather API
type OneCallOverviewResponse struct {
	Lat             float64 `json:"lat"`
	Lon             float64 `json:"lon"`
	TZ              string  `json:"tz"`
	Date            string  `json:"date"`
	Units           string  `json:"units"`
	WeatherOverview string  `json:"weather_overview"`
}

// Struct to represent the current weather response from OpenWeather API
type CurrentWeatherResponse struct {
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Humidity  int     `json:"humidity"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
	} `json:"main"`
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
	Name string `json:"name"`
}

// Struct to represent the weather overview information to be sent to the frontend for display in the weather widget
type WeatherOverviewResponse struct {
	Location    string `json:"location"`
	Temp        int    `json:"temp"`
	FeelsLike   int    `json:"feels_like"`
	TempMin     int    `json:"temp_min"`
	TempMax     int    `json:"temp_max"`
	Humidity    int    `json:"humidity"`
	WindSpeed   int    `json:"wind_speed"`
	Condition   string `json:"condition"`
	Description string `json:"description"`
	Message     string `json:"message"`
}

func NewWeatherService(apiKey string) *WeatherService {
	return &WeatherService{
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: OpenWeatherBaseURL,
	}
}

/*
Uses OpenWeather API to get coordinates from a US zip code, then fetches current weather
and the overview for the day. Used with a weather screen widget's config.zip.
The day's overview text is optional: if that request fails (it needs a One Call 3.0
subscription), the current weather is still returned with an empty Message.
zip - 5 digit US zip code
*/
func (s *WeatherService) GetWeatherForZip(zip string) (*WeatherOverviewResponse, error) {
	if zip == "" {
		return nil, errors.New("zip cannot be empty")
	}

	// 1. Geocode zipcode to coordinates
	url := fmt.Sprintf(
		"%s/geo/1.0/zip?zip=%s,US&appid=%s",
		s.baseURL,
		neturl.QueryEscape(zip),
		s.apiKey,
	)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed geocode request: %w", redactURLErr(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("geocode API returned status %d", resp.StatusCode)
	}

	var coordinate_response CoordinateResponse
	if err := json.NewDecoder(resp.Body).Decode(&coordinate_response); err != nil {
		return nil, fmt.Errorf("failed parsing geocode response: %v", err)
	}

	unitsParam := "imperial"

	// 2. Fetch current weather (temp, conditions, etc.)
	currentURL := fmt.Sprintf(
		"%s/data/2.5/weather?lat=%f&lon=%f&units=%s&appid=%s",
		s.baseURL, coordinate_response.Lat, coordinate_response.Lon, unitsParam, s.apiKey,
	)

	currentResp, err := s.client.Get(currentURL)
	if err != nil {
		return nil, fmt.Errorf("failed current weather request: %w", redactURLErr(err))
	}
	defer currentResp.Body.Close()

	if currentResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("current weather API returned status %d", currentResp.StatusCode)
	}

	var currentWeather CurrentWeatherResponse
	if err := json.NewDecoder(currentResp.Body).Decode(&currentWeather); err != nil {
		return nil, fmt.Errorf("failed parsing current weather response: %v", err)
	}

	// 3. Fetch overview text (optional; the widget still works without it)
	overviewURL := fmt.Sprintf(
		"%s/data/3.0/onecall/overview?lon=%f&lat=%f&units=%s&appid=%s",
		s.baseURL, coordinate_response.Lon, coordinate_response.Lat, unitsParam, s.apiKey,
	)
	oneCallResponse, err := s.getOverview(overviewURL)
	if err != nil {
		log.Printf("weather: no daily overview for %s: %v", zip, err)
		oneCallResponse = &OneCallOverviewResponse{}
	}

	// 4. Build structured response
	condition := ""
	description := ""
	if len(currentWeather.Weather) > 0 {
		condition = currentWeather.Weather[0].Main
		description = currentWeather.Weather[0].Description
	}

	location := currentWeather.Name
	if location == "" {
		location = coordinate_response.Name
	}

	return &WeatherOverviewResponse{
		Location:    location,
		Temp:        int(math.Round(currentWeather.Main.Temp)),
		FeelsLike:   int(math.Round(currentWeather.Main.FeelsLike)),
		TempMin:     int(math.Round(currentWeather.Main.TempMin)),
		TempMax:     int(math.Round(currentWeather.Main.TempMax)),
		Humidity:    currentWeather.Main.Humidity,
		WindSpeed:   int(math.Round(currentWeather.Wind.Speed)),
		Condition:   condition,
		Description: description,
		Message:     oneCallResponse.WeatherOverview,
	}, nil
}

// redactURLErr removes request URLs, which contain the OpenWeather API key.
func redactURLErr(err error) error {
	var urlErr *neturl.Error
	if errors.As(err, &urlErr) {
		return urlErr.Err
	}
	return err
}

// getOverview fetches the day's weather overview text
func (s *WeatherService) getOverview(url string) (*OneCallOverviewResponse, error) {
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed overview request: %w", redactURLErr(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("overview API returned status %d", resp.StatusCode)
	}

	var oneCallResponse OneCallOverviewResponse
	if err := json.NewDecoder(resp.Body).Decode(&oneCallResponse); err != nil {
		return nil, fmt.Errorf("failed parsing overview response: %v", err)
	}
	return &oneCallResponse, nil
}
