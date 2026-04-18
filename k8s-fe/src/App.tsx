// App.tsx
import { useEffect, useState } from "react";
import { createJob, getJobs } from "./api";

function App() {
  const [jobs, setJobs] = useState([]);
  const [name, setName] = useState("");

  async function loadJobs() {
    const data = await getJobs();
    setJobs(data);
  }

  async function handleSubmit() {
    await createJob(name);
    setName("");
    loadJobs();
  }

  useEffect(() => {
    loadJobs();
  }, []);

  return (
    <div>
      <h1>Job Queue</h1>

      <input
        value={name}
        onChange={(e) => setName(e.target.value)}
        placeholder="Job name"
      />
      <button onClick={handleSubmit}>Create Job</button>

      <ul>
        {jobs.map((job: any) => (
          <li key={job.id}>
            {job.name} - {job.status}
          </li>
        ))}
      </ul>
    </div>
  );
}

export default App;