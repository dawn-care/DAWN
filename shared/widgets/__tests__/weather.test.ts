import { CHILLY_BELOW_F, WARM_FROM_F, weatherWord } from "../weather";

describe("weatherWord", () => {
  it("says Chilly for the mockup's 48 degrees and cloud", () => {
    expect(weatherWord({ temp: 48, condition: "Clouds" })).toBe("Chilly");
  });

  it.each([
    ["Thunderstorm", "Stormy"],
    ["Rain", "Rainy"],
    ["Drizzle", "Rainy"],
    ["Snow", "Snowy"],
  ])("says the %s word however warm or cold it is", (condition, expected) => {
    for (const temp of [10, 50, 70, 95]) {
      expect(weatherWord({ temp, condition })).toBe(expected);
    }
  });

  it("is Chilly below the chilly line and not at it", () => {
    expect(weatherWord({ temp: CHILLY_BELOW_F - 1, condition: "Clear" })).toBe("Chilly");
    expect(weatherWord({ temp: CHILLY_BELOW_F, condition: "Clear" })).toBe("Sunny");
  });

  it("is Warm from the warm line and not below it", () => {
    expect(weatherWord({ temp: WARM_FROM_F, condition: "Clear" })).toBe("Warm");
    expect(weatherWord({ temp: WARM_FROM_F - 1, condition: "Clear" })).toBe("Sunny");
  });

  it("says Warm and Chilly whatever the sky", () => {
    expect(weatherWord({ temp: 85, condition: "Clouds" })).toBe("Warm");
    expect(weatherWord({ temp: 30, condition: "Clear" })).toBe("Chilly");
  });

  it("says Sunny in mild clear weather and Cloudy in mild cloud", () => {
    expect(weatherWord({ temp: 65, condition: "Clear" })).toBe("Sunny");
    expect(weatherWord({ temp: 65, condition: "Clouds" })).toBe("Cloudy");
  });

  it.each(["Mist", "Fog", "Haze", "Smoke"])("says Cloudy for %s in mild weather", (condition) => {
    expect(weatherWord({ temp: 65, condition })).toBe("Cloudy");
  });

  it("ignores case and spaces in the condition", () => {
    expect(weatherWord({ temp: 65, condition: "  RAIN " })).toBe("Rainy");
    expect(weatherWord({ temp: 65, condition: "clear" })).toBe("Sunny");
  });

  it("goes by how it feels, not the air temperature", () => {
    expect(weatherWord({ temp: 60, feels_like: 48, condition: "Clear" })).toBe("Chilly");
    expect(weatherWord({ temp: 70, feels_like: 78, condition: "Clear" })).toBe("Warm");
    expect(weatherWord({ temp: 50, feels_like: 60, condition: "Clear" })).toBe("Sunny");
    expect(weatherWord({ temp: 80, feels_like: 70, condition: "Clouds" })).toBe("Cloudy");
  });

  it("applies the same lines to the feels-like temperature", () => {
    expect(weatherWord({ temp: 65, feels_like: CHILLY_BELOW_F - 1, condition: "Clear" })).toBe("Chilly");
    expect(weatherWord({ temp: 65, feels_like: CHILLY_BELOW_F, condition: "Clear" })).toBe("Sunny");
    expect(weatherWord({ temp: 65, feels_like: WARM_FROM_F, condition: "Clear" })).toBe("Warm");
  });

  it("matches the mockup when it feels colder than 48 degrees", () => {
    expect(weatherWord({ temp: 48, feels_like: 44, condition: "Clouds" })).toBe("Chilly");
  });

  it("still lets rain, snow and storms win over how it feels", () => {
    expect(weatherWord({ temp: 80, feels_like: 85, condition: "Rain" })).toBe("Rainy");
    expect(weatherWord({ temp: 20, feels_like: 5, condition: "Snow" })).toBe("Snowy");
  });

  it("falls back to the air temperature when feels-like is missing or not a number", () => {
    expect(weatherWord({ temp: 40, condition: "Clear" })).toBe("Chilly");
    expect(weatherWord({ temp: 40, feels_like: Number.NaN, condition: "Clear" })).toBe("Chilly");
    expect(weatherWord({ temp: 90, feels_like: undefined, condition: "Clear" })).toBe("Warm");
  });

  it("takes the weather widget's data as it is", () => {
    const data = {
      location: "Irondequoit",
      temp: 48,
      feels_like: 44,
      temp_min: 45,
      temp_max: 51,
      humidity: 71,
      wind_speed: 8,
      condition: "Clouds",
      description: "overcast clouds",
      message: "",
    };
    expect(weatherWord(data)).toBe("Chilly");
  });

  it("uses the feels-like temperature when the air temperature is missing", () => {
    expect(weatherWord({ temp: Number.NaN, feels_like: 40, condition: "Clear" })).toBe("Chilly");
  });

  it("uses the condition alone when the temperature is missing", () => {
    expect(weatherWord({ temp: Number.NaN, condition: "Clear" })).toBe("Sunny");
    expect(weatherWord({ temp: Number.NaN, condition: "Snow" })).toBe("Snowy");
  });

  it("uses the temperature alone when the condition is missing", () => {
    expect(weatherWord({ temp: 40, condition: "" })).toBe("Chilly");
    expect(weatherWord({ temp: 90, condition: "" })).toBe("Warm");
  });

  it("has no word when there is nothing to go on", () => {
    expect(weatherWord({ temp: Number.NaN, condition: "" })).toBeNull();
    expect(weatherWord({ temp: 65, condition: "" })).toBeNull();
  });
});

describe("weatherWord with a missing condition", () => {
  it("falls back to the temperature instead of throwing", () => {
    expect(weatherWord({ temp: 40 })).toBe("Chilly");
    expect(weatherWord({ temp: 80, condition: null })).toBe("Warm");
  });

  it("is null when there is nothing to go on", () => {
    expect(weatherWord({ temp: 65 })).toBeNull();
  });
});
