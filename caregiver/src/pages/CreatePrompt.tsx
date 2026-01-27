import { useState } from "react";

type DisplayLength = "1 Minute" | "5 Minutes" | "15 Minutes";
type RepeatType = "daily" | "weekly" | "monthly" | "yearly" | "never";

const allowedExtensions = [".jpg", ".jpeg", ".png", ".heic"];

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

export default function CreatePrompt() {
  const [promptMessage, setPromptMessage] = useState("");
  const [displayLength, setDisplayLength] = useState<DisplayLength>("1 Minute");
  const [repeat, setRepeat] = useState<RepeatType>("daily");
  const [weeklyDay, setWeeklyDay] = useState("Monday");
  const [monthlyDay, setMonthlyDay] = useState<number>(1);
  const [yearlyDate, setYearlyDate] = useState("2025-01-01");
  const [file, setFile] = useState<File | null>(null);
  const [fileError, setFileError] = useState("");
  const [uploadStatus, setUploadStatus] = useState("");

  function validateFile(f: File) {
    const ext = "." + f.name.split(".").pop()?.toLowerCase();
    if (!allowedExtensions.includes(ext)) {
      setFileError(`Invalid file type for prompts.`);
      setFile(null);
    } else {
      setFileError("");
      setFile(f);
    }
  }

  async function handleUpload() {
    if (!promptMessage.trim()) {
      setUploadStatus("Please enter text to go alongside this picture.");
      return;
    }

    try {
      let imageUri: string | null = null;

      // If image is included, upload it to the photo library first
      if (file) {
        const formData = new FormData();
        formData.append("file", file, file.name);

        const imageResponse = await fetch(`${API_URL}/photo/upload`, {
            method: "POST",
            body: formData,
        });

        if (!imageResponse.ok) {
            const text = await imageResponse.text();
            throw new Error(text || "Image upload failed");
        }

        // After image is uploaded, get the URI
        const imageData = await imageResponse.json();
        imageUri = imageData.uri;
      }

      // Create the prompt 
      const promptResponse = await fetch(`${API_URL}/messages`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          message: promptMessage,
          image_uri: imageUri,
          /*
          NOTE 1: This currently works for message prompts. But do we really need different models for caregiver/partner locations?
                  Seems like the locations could be easily included in the message text and a picture of a map can be included without needing a separate model

          NOTE 2: Repeat is not currently in the prompt model, should it be?

          repeat,
            weeklyDay: repeat === "weekly" ? weeklyDay : undefined,
            monthlyDay: repeat === "monthly" ? monthlyDay : undefined,
            yearlyDate: repeat === "yearly" ? yearlyDate : undefined,
          */
        }),
      });

      if (!promptResponse.ok) {
        const text = await promptResponse.text();
        throw new Error(text || "Message creation failed");
      }

      // Success!
      setUploadStatus("Saved!");
      setFile(null);
      setPromptMessage("");
    } catch (err: any) {
      setUploadStatus("Upload failed: " + err.message);
    }
  }

  return (
    <div className="w-full flex flex-col items-center mt-8 px-6">
      <h1 className="text-3xl font-semibold mb-6">Create Message</h1>
      <div className="border p-10 max-w-3xl w-full rounded-lg bg-white shadow">

        {/* Prompt Message */}
        <div className="mb-6">
          <label className="text-lg font-medium">Message Text:</label>
          <textarea
            value={promptMessage}
            maxLength={300}
            onChange={(e) => setPromptMessage(e.target.value)}
            className="w-full mt-2 p-3 border rounded bg-purple-50"
            placeholder="Enter text to display on the care partners screen (300 character limit)"
          />
        </div>

        {/* Prompt Display Length */}
        <div className="mb-6">
          <label className="text-lg font-medium">Shows For:</label>
          <select
            value={displayLength}
            onChange={(e) => setDisplayLength(e.target.value as DisplayLength)}
            className="w-full mt-2 p-3 border rounded bg-purple-50"
          >
            <option value="1 Minute">1 Minute</option>
            <option value="5 Minutes">5 Minutes</option>
            <option value="15 Minutes">15 Minutes</option>
          </select>
        </div>

        {/* Repeat Option */}
        <div className="mb-6">
          <label className="text-lg font-medium">Repeat:</label>
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

        {/* Conditional Repeat Inputs */}
        {repeat === "weekly" && (
          <div className="mb-6">
            <label className="text-lg font-medium">Select Weekday:</label>
            <select
              value={weeklyDay}
              onChange={(e) => setWeeklyDay(e.target.value)}
              className="w-full mt-2 p-3 border rounded bg-purple-50"
            >
              {["Monday","Tuesday","Wednesday","Thursday","Friday","Saturday","Sunday"].map(day => (
                <option key={day} value={day}>{day}</option>
              ))}
            </select>
          </div>
        )}

        {repeat === "monthly" && (
          <div className="mb-6">
            <label className="text-lg font-medium">Select Day:</label>
            <select
              value={monthlyDay}
              onChange={(e) => setMonthlyDay(parseInt(e.target.value))}
              className="w-full mt-2 p-3 border rounded bg-purple-50"
            >
              {Array.from({length:28}, (_,i)=>i+1).map(day => (
                <option key={day} value={day}>{day}</option>
              ))}
            </select>
          </div>
        )}

        {repeat === "yearly" && (
          <div className="mb-6">
            <label className="text-lg font-medium">Select Date:</label>
            <input
              type="date"
              value={yearlyDate}
              onChange={(e) => setYearlyDate(e.target.value)}
              className="w-full mt-2 p-3 border rounded bg-purple-50"
            />
          </div>
        )}

        {/* Optional Picture to Include */}
        <div className="mb-6">
          <label className="text-lg font-medium">Picture to Display Above Text (Optional):</label>
          <div
            className="mt-2 p-6 border-2 border-dashed border-gray-400 rounded cursor-pointer bg-purple-50 text-center"
            onClick={() => document.getElementById("fileInput")?.click()}
          >
            {file ? <span>{file.name}</span> : <span>Drag and drop or click to upload</span>}
          </div>

          <input
            id="fileInput"
            type="file"
            className="hidden"
            onChange={(e) => {
              const f = e.target.files?.[0];
              f && validateFile(f);
            }}
          />

          {fileError && <p className="text-red-500 mt-2 text-sm">{fileError}</p>}
        </div>

        {/* Save */}
        <button
          className="mt-4 w-full bg-teal-200 hover:bg-teal-300 text-gray-800 font-semibold py-3 rounded"
          onClick={handleUpload}
        >
          Save
        </button>

        {uploadStatus && <p className="mt-2 text-sm">{uploadStatus}</p>}
      </div>
    </div>
  );
}
