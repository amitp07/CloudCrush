// src/api.ts
const BASE_URL = "http://localhost:8080";

export async function createJob(name: string) {
  await fetch(`${BASE_URL}/jobs`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name }),
  });
}

export async function getJobs() {
  const res = await fetch(`${BASE_URL}/jobs`);
  return res.json();
}