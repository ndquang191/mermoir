import { Photo } from "../api/client";

interface PhotoGridProps {
  photos: Photo[];
}

export function PhotoGrid({ photos }: PhotoGridProps) {
  return (
    <div
      style={{
        display: "grid",
        gridTemplateColumns: "repeat(auto-fill, minmax(120px, 1fr))",
        gap: "0.5rem",
      }}
    >
      {photos.map((photo) => (
        <PhotoCell key={photo.id} photo={photo} />
      ))}
    </div>
  );
}

function PhotoCell({ photo }: { photo: Photo }) {
  if (photo.status === "ready" && photo.thumb_path) {
    return (
      <img
        src={`/storage/thumb/${photo.id}.jpg`}
        alt="A moment captured in this memory"
        style={{
          width: "100%",
          aspectRatio: "1",
          objectFit: "cover",
          borderRadius: 4,
          display: "block",
        }}
      />
    );
  }

  return (
    <div
      style={{
        width: "100%",
        aspectRatio: "1",
        background: "#f0f0f0",
        borderRadius: 4,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        fontSize: "0.75rem",
        color: "#999",
      }}
    >
      {photo.status === "failed" ? "Failed" : "Processing..."}
    </div>
  );
}
