/**
 * weather.ts
 *
 * The one plain word the weather tile shows, such as "Chilly". The words and
 * the temperature bands below are provisional until they are agreed.
 */

/** Below this feels-like temperature (degrees Fahrenheit) the word is "Chilly". Provisional. */
export const CHILLY_BELOW_F = 55;

/** From this feels-like temperature (degrees Fahrenheit) the word is "Warm". Provisional. */
export const WARM_FROM_F = 75;

export type WeatherWord = "Stormy" | "Rainy" | "Snowy" | "Chilly" | "Warm" | "Sunny" | "Cloudy";

/**
 * What the weather lookup reports, with the server's own field names so the
 * weather widget's data can be passed in as it is. Temperatures are in degrees
 * Fahrenheit.
 */
export type WeatherReading = {
  temp: number;
  /** What it feels like, which already counts wind and humidity. Used in place of `temp` when present. */
  feels_like?: number;
  /** The main condition, such as "Clouds", "Rain" or "Clear". Case does not matter. */
  condition: string;
};

const STORM = ["thunderstorm", "squall", "tornado"];
const RAIN = ["rain", "drizzle"];
const SNOW = ["snow"];
const CLEAR = ["clear"];

/**
 * The word for the weather: rain, snow and storms come first, then chilly or
 * warm by how it feels outside, then the sky. Null when there is nothing to go on.
 */
export function weatherWord(reading: WeatherReading): WeatherWord | null {
  const condition = reading.condition.trim().toLowerCase();
  if (STORM.includes(condition)) return "Stormy";
  if (RAIN.includes(condition)) return "Rainy";
  if (SNOW.includes(condition)) return "Snowy";

  const felt = Number.isFinite(reading.feels_like) ? reading.feels_like : reading.temp;
  if (felt !== undefined && Number.isFinite(felt)) {
    if (felt < CHILLY_BELOW_F) return "Chilly";
    if (felt >= WARM_FROM_F) return "Warm";
  }

  if (CLEAR.includes(condition)) return "Sunny";
  // Clouds, and anything else that hides the sky such as mist, fog or haze.
  if (condition !== "") return "Cloudy";
  return null;
}
