import { useState } from "react";

type MediaType = "video" | "photo" | "audio";
type RepeatType = "daily" | "weekly" | "monthly" | "yearly" | "never";

const allowedExtensions = {
  video: [".mp4", ".mov", ".avi"],
  photo: [".jpg", ".jpeg", ".png", ".heic"],
  audio: [".mp3", ".wav", ".ogg"],
};

export default function UploadMedia() {
  const [mediaName, setMediaName] = useState("");
  const [mediaType, setMediaType] = useState<MediaType>("video");
  const [repeat, setRepeat] = useState<RepeatType>("daily");
  const [weeklyDay, setWeeklyDay] = useState("");
  const [monthlyDay, setMonthlyDay] = useState<number | null>(null);
  const [yearlyDate, setYearlyDate] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [fileError, setFileError] = useState("");

  function validateFile(f: File) {
    const ext = "." + f.name.split(".").pop()?.toLowerCase();
    if (!allowedExtensions[mediaType].includes(ext)) {
      setFileError(`Invalid file type for ${mediaType.toUpperCase()}.`);
      setFile(null);
    } else {
      setFileError("");
      setFile(f);
    }
  }

  return (
    <div className="w-full flex flex-col items-center mt-8 px-6">
      <h1 className="text-3xl font-semibold mb-6">Add Media</h1>

      <div className="border p-10 max-w-3xl w-full rounded-lg bg-white shadow">
        {/* Media Name */}
        <div className="mb-6">
          <label className="text-lg font-medium">Media Name:</label>
          <input
            type="text"
            value={mediaName}
            onChange={(e) => setMediaName(e.target.value)}
            className="w-full mt-2 p-3 border rounded bg-purple-50"
            placeholder="Enter media name"
          />
        </div>

        {/* Media Type */}
        <div className="mb-6">
          <label className="text-lg font-medium">Media Type:</label>
          <select
            value={mediaType}
            onChange={(e) => setMediaType(e.target.value as MediaType)}
            className="w-full mt-2 p-3 border rounded bg-purple-50"
          >
            <option value="video">Video</option>
            <option value="photo">Photo</option>
            <option value="audio">Audio</option>
          </select>
        </div>

        {/* Upload Media */}
        <div className="mb-6">
          <label className="text-lg font-medium">Upload Media:</label>

          <div
            className="mt-2 p-6 border-2 border-dashed border-gray-400 rounded cursor-pointer bg-purple-50 text-center"
            onClick={() => document.getElementById("fileInput")?.click()}
          >
            {file ? (
              <span className="font-medium">{file.name}</span>
            ) : (
              <span className="text-gray-500">
                Drag and drop or click to upload
              </span>
            )}
          </div>

          <input
            id="fileInput"
            type="file"
            className="hidden"
            onChange={(e) => {
              const f = e.target.files?.[0];
              if (f) validateFile(f);
            }}
          />

          {fileError && (
            <p className="text-red-500 mt-2 text-sm">{fileError}</p>
          )}
        </div>

        {/* Repeats */}
        <div className="mb-6">
          <label className="text-lg font-medium">Repeats:</label>
          <select
            value={repeat}
            onChange={(e) => setRepeat(e.target.value as RepeatType)}
            className="w-full mt-2 p-3 border rounded bg-purple-50"
          >
            <option value="daily">Daily</option>
            <option value="weekly">Weekly</option>
            <option value="monthly">Monthly</option>
            <option value="yearly">Yearly</option>
            <option value="never">Never</option>
          </select>
        </div>

        {/* Conditional Fields */}
        {repeat === "weekly" && (
          <div className="mb-6">
            <label className="text-lg font-medium">Every:</label>
            <select
              value={weeklyDay}
              onChange={(e) => setWeeklyDay(e.target.value)}
              className="w-full mt-2 p-3 border rounded bg-purple-50"
            >
              <option value="">Select a weekday</option>
              {["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"].map(
                (day) => (
                  <option key={day} value={day.toLowerCase()}>
                    {day}
                  </option>
                )
              )}
            </select>
          </div>
        )}

        {repeat === "monthly" && (
          <div className="mb-6">
            <label className="text-lg font-medium">Day of Month:</label>
            <select
              value={monthlyDay ?? ""}
              onChange={(e) => setMonthlyDay(Number(e.target.value))}
              className="w-full mt-2 p-3 border rounded bg-purple-50"
            >
              <option value="">Select day</option>
              {Array.from({ length: 29 }, (_, i) => i + 1).map((n) => (
                <option key={n} value={n}>
                  {n}
                </option>
              ))}
            </select>
          </div>
        )}

        {repeat === "yearly" && (
          <div className="mb-6">
            <label className="text-lg font-medium">Date:</label>
            <input
              type="date"
              value={yearlyDate}
              onChange={(e) => setYearlyDate(e.target.value)}
              className="w-full mt-2 p-3 border rounded bg-purple-50"
            />
          </div>
        )}

        {/* Save Button */}
        <button
          className="mt-4 w-full bg-teal-200 hover:bg-teal-300 text-gray-800 font-semibold py-3 rounded"
          onClick={() => alert("Saved!")}
        >
          Save
        </button>
      </div>
    </div>
  );
}
