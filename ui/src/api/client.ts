import axios from "axios";

const http = axios.create({
  baseURL: "/api",
  headers: { "Content-Type": "application/json" },
});

export interface Photo {
  id: string;
  entry_id: string;
  raw_path: string;
  thumb_path: string;
  status: "pending" | "ready" | "failed";
}

export interface Entry {
  id: string;
  date: string;
  story: string;
  created_at: string;
  photos: Photo[];
}

export interface CreateEntryInput {
  date: string;
  story: string;
}

export async function getEntries(): Promise<Entry[]> {
  const res = await http.get<Entry[]>("/entries");
  return res.data;
}

export async function createEntry(input: CreateEntryInput): Promise<Entry> {
  const res = await http.post<Entry>("/entries", input);
  return res.data;
}

export async function uploadPhoto(entryId: string, file: File): Promise<Photo> {
  const form = new FormData();
  form.append("photo", file);
  const res = await axios.post<Photo>(`/api/entries/${entryId}/photos`, form, {
    headers: { "Content-Type": "multipart/form-data" },
  });
  return res.data;
}
