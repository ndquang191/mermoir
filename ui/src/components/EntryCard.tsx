import { Entry } from "../api/client";
import { PhotoGrid } from "./PhotoGrid";

interface EntryCardProps {
  entry: Entry;
}

export function EntryCard({ entry }: EntryCardProps) {
  const snippet =
    entry.story.length > 200 ? entry.story.slice(0, 200) + "..." : entry.story;

  return (
    <article
      style={{
        border: "1px solid #eee",
        borderRadius: 8,
        padding: "1.25rem",
        marginBottom: "1.5rem",
      }}
    >
      <time
        dateTime={entry.date}
        style={{ fontSize: "0.85rem", color: "#888", display: "block", marginBottom: "0.5rem" }}
      >
        {new Date(entry.date + "T00:00:00").toLocaleDateString("en-US", {
          year: "numeric",
          month: "long",
          day: "numeric",
        })}
      </time>

      {snippet && (
        <p style={{ margin: "0 0 1rem", lineHeight: 1.6 }}>{snippet}</p>
      )}

      {entry.photos.length > 0 && <PhotoGrid photos={entry.photos} />}
    </article>
  );
}
