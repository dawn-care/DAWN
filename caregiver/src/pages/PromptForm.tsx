import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

const allowedExtensions = [".jpg", ".jpeg", ".png", ".heic"];

// Create a prop interface so that Library can pass parameters to the form for editing existing prompts without relying on URL params
interface PromptFormProps {
  moduleId?: number;
  onDeleteSuccess?: () => void;
  onUpdateSuccess?: () => void;
}

const API_URL = import.meta.env.VITE_API_URL;

export default function PromptForm({
        moduleId: propModuleId,
        onDeleteSuccess,
        onUpdateSuccess
    }: PromptFormProps
  ) {
  // Creating Variables
  const [promptMessage, setPromptMessage] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [fileError, setFileError] = useState("");
  const [uploadStatus, setUploadStatus] = useState("");
  // Editing Variables
  const { moduleId } = useParams<{ moduleId: string }>();
  const actualModuleId = propModuleId ?? moduleId;
  const [promptId, setPromptId] = useState<number | null>(null);
  const isEditMode = Boolean(actualModuleId);
  const [existingImageUri, setExistingImageUri] = useState<string | null>(null);
  // Deleting Variables
  const navigate = useNavigate();
  const { user, activeCarePartnerId } = useAuth();

  /*******************************************
   * If form is in edit mode, load message
   *******************************************/
  useEffect(() => {
    if (!actualModuleId) return;

    async function loadPrompt() {
      var res = await fetch(`${API_URL}/modules/${actualModuleId}`);
      const module = await res.json();
      console.log("Loaded module for editing:", module);
      setPromptId(module.metadata_id);

      fetch(`${API_URL}/messages?id=${module.metadata_id}`)
        .then(res => {
          if (!res.ok) throw new Error("Failed to load message");
          return res.json();
        })
        .then(data => {
          setPromptMessage(data.message);
          setExistingImageUri(data.image_uri ?? null);
        })
        .catch(err => setUploadStatus(err.message));
    }

    loadPrompt();
  }, [actualModuleId]);

  /*******************************************
   * Check if the uploaded file is a picture
   *******************************************/ 
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

  /*******************************************
   * Message Creation and Photo Upload
   *******************************************/ 
  async function handleUpload() {
    if (!promptMessage.trim()) {
      setUploadStatus("Please enter text to go alongside this picture.");
      return;
    }

    try {
      let imageUri: string | null = null;
      let metadataId: number | null = null;

      // Upload the optional image first, if provided
      if (file) {
        if (!user?.id || !activeCarePartnerId) {
          setUploadStatus("Cannot upload: user information missing.");
          return;
        }

        const formData = new FormData();
        formData.append("file", file, file.name);
        formData.append("caregiver_id", user.id.toString());
        formData.append("carepartner_id", activeCarePartnerId.toString());

        const imageResponse = await fetch(`${API_URL}/photo/upload`, {
          method: "POST",
          body: formData,
        });

        if (!imageResponse.ok) {
          const text = await imageResponse.text();
          throw new Error(text || "Image upload failed");
        }

        const imageData = await imageResponse.json();
        var uri = `/photo/file?id=${imageData.id}`;
        imageUri = uri;
        metadataId = imageData.id; // assuming the image upload returns a metadata id
      }

      // Create the message/prompt
      const promptResponse = await fetch(`${API_URL}/messages`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          message: promptMessage,
          image_uri: imageUri,
        }),
      });

      if (!promptResponse.ok) {
        const text = await promptResponse.text();
        throw new Error(text || "Message creation failed");
      }

      const promptData = await promptResponse.json();
      // Use the returned id from the prompt creation as the metadata ID
      if (!metadataId) metadataId = promptData.id;

      if (!metadataId) {
        setUploadStatus("Saved but could not determine metadata ID.");
        return;
      }

      // Create a module pointing to this metadata
      const moduleResponse = await fetch(`${API_URL}/modules`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          module_type: "messages",
          caregiver_id: user?.id,
          carepartner_id: activeCarePartnerId,
          metadata_id: metadataId,
        }),
      });

      if (!moduleResponse.ok) {
        const text = await moduleResponse.text();
        setUploadStatus(`Prompt saved but failed to create module: ${text}`);
        return;
      }

      setUploadStatus("Saved!");
      setFile(null);
      setPromptMessage("");
    } catch (err: any) {
      setUploadStatus("Upload failed: " + err.message);
    }
  }

  /*******************************************
   * Update/Edit
   *******************************************/ 
  async function handleUpdate() {
    if (!promptMessage.trim()) {
      setUploadStatus("Message text cannot be empty.");
      return;
    }

    try {
      let imageUri = existingImageUri;

      if (file) {
        const formData = new FormData();
        formData.append("file", file);

        const res = await fetch(`${API_URL}/photo/upload`, {
          method: "POST",
          body: formData,
        });

        if (!res.ok) throw new Error("Image upload failed");
        imageUri = (await res.json()).uri;
      }

      const res = await fetch(`${API_URL}/messages/update/${promptId}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          message: promptMessage,
          image_uri: imageUri,
        }),
      });

      if (!res.ok) throw new Error("Update failed");

      setUploadStatus("Updated");
      if (onUpdateSuccess) {
          onUpdateSuccess();
      }
    } catch (err: any) {
      setUploadStatus(err.message);
    }
  }

  /*******************************************
   * Delete
   *******************************************/ 
  async function handleDelete() {
    if (!confirm("Are you sure you want to delete this message?")) return;

    try {
      const moduleRes = await fetch(`${API_URL}/modules/delete?id=${actualModuleId}`, {
        method: "DELETE",
      });

      if (!moduleRes.ok) throw new Error("Delete failed");

      const promptRes = await fetch(`${API_URL}/messages/delete/${promptId}`, {
        method: "DELETE",
      });

      if (!promptRes.ok) throw new Error("Delete failed");

      setUploadStatus("Deleted");
      if (onDeleteSuccess) {
          onDeleteSuccess();
      } else {
          navigate("/library");
      }
    } catch (err: any) {
      setUploadStatus(err.message);
    }
  }

  /*******************************************
   * Returned Webpage
   *******************************************/ 
  return (
    <div className="w-full flex flex-col items-center mt-8 px-6">
      <h1 className="text-3xl font-semibold mb-6">
        {isEditMode ? "Edit Message" : "Create Message"}
      </h1>
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

        {isEditMode ? (
          <>
            {/* Edit Mode */}
              <button
                className="mt-4 w-full bg-teal-200 hover:bg-teal-300 text-gray-800 font-semibold py-3 rounded"
                onClick={handleUpdate}
              >
                Update
              </button>
                        
              <button
                className="mt-4 w-full bg-teal-200 hover:bg-teal-300 text-gray-800 font-semibold py-3 rounded"
                onClick={handleDelete}
              >
                Delete
              </button>          
          </>
          ) : (
          <>
            {/* Create Mode */}
              <button
                className="mt-4 w-full bg-teal-200 hover:bg-teal-300 text-gray-800 font-semibold py-3 rounded"
                onClick={handleUpload}
              >
                Save
              </button>

              {uploadStatus && <p className="mt-2 text-sm">{uploadStatus}</p>}
          </>
        )}
      </div>
    </div>
  );
}
