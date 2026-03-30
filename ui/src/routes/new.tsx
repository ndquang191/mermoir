import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useState, FormEvent } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createEntry, uploadPhoto } from "../api/client";

export const Route = createFileRoute("/new")({
  component: NewEntryPage,
});

function NewEntryPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [date, setDate] = useState(new Date().toISOString().split("T")[0] ?? "");
  const [story, setStory] = useState("");
  const [files, setFiles] = useState<FileList | null>(null);
  const [error, setError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: async () => {
      const entry = await createEntry({ date, story });

      if (files && files.length > 0) {
        const uploads = Array.from(files).map((file) =>
          uploadPhoto(entry.id, file)
        );
        await Promise.all(uploads);
      }

      return entry;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["entries"] });
      navigate({ to: "/" });
    },
    onError: (err: Error) => {
      setError(err.message);
    },
  });

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    mutation.mutate();
  };

  return (
    <div style={{ maxWidth: 600 }}>
      <h2>New Memory</h2>
      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: "1rem" }}>
          <label htmlFor="date" style={{ display: "block", marginBottom: 4 }}>
            Date
          </label>
          <input
            id="date"
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
            required
            style={{ padding: "0.5rem", width: "100%", boxSizing: "border-box" }}
          />
        </div>

        <div style={{ marginBottom: "1rem" }}>
          <label htmlFor="story" style={{ display: "block", marginBottom: 4 }}>
            Story
          </label>
          <textarea
            id="story"
            value={story}
            onChange={(e) => setStory(e.target.value)}
            rows={6}
            placeholder="Write about this memory..."
            style={{ padding: "0.5rem", width: "100%", boxSizing: "border-box" }}
          />
        </div>

        <div style={{ marginBottom: "1.5rem" }}>
          <label htmlFor="photos" style={{ display: "block", marginBottom: 4 }}>
            Photos
          </label>
          <input
            id="photos"
            type="file"
            accept="image/*"
            multiple
            onChange={(e) => setFiles(e.target.files)}
            style={{ display: "block" }}
          />
        </div>

        {error && (
          <p style={{ color: "red", marginBottom: "1rem" }}>{error}</p>
        )}

        <button
          type="submit"
          disabled={mutation.isPending}
          style={{
            padding: "0.6rem 1.5rem",
            background: "#333",
            color: "#fff",
            border: "none",
            cursor: "pointer",
            borderRadius: 4,
          }}
        >
          {mutation.isPending ? "Saving..." : "Save Memory"}
        </button>
      </form>
    </div>
  );
}
