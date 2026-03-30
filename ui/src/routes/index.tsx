import { createFileRoute } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { getEntries } from "../api/client";
import { EntryCard } from "../components/EntryCard";

export const Route = createFileRoute("/")({
  component: TimelinePage,
});

function TimelinePage() {
  const { data: entries, isLoading, isError } = useQuery({
    queryKey: ["entries"],
    queryFn: getEntries,
  });

  if (isLoading) {
    return <p>Loading memories...</p>;
  }

  if (isError) {
    return <p style={{ color: "red" }}>Failed to load entries. Is the API running?</p>;
  }

  if (!entries || entries.length === 0) {
    return (
      <div style={{ textAlign: "center", color: "#888", marginTop: "4rem" }}>
        <p>No memories yet.</p>
        <a href="/new">Create your first entry</a>
      </div>
    );
  }

  return (
    <div>
      <h2>Our Timeline</h2>
      {entries.map((entry) => (
        <EntryCard key={entry.id} entry={entry} />
      ))}
    </div>
  );
}
