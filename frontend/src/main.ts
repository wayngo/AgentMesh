import { createRun, watchRun } from "./api";
import "./styles.css";

const form = document.querySelector<HTMLFormElement>("#run-form")!;
const agentInput = document.querySelector<HTMLInputElement>("#agent-id")!;
const promptInput = document.querySelector<HTMLTextAreaElement>("#input")!;
const submit = document.querySelector<HTMLButtonElement>("#submit")!;
const status = document.querySelector<HTMLElement>("#status")!;
const result = document.querySelector<HTMLElement>("#result")!;

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  submit.disabled = true; result.textContent = ""; status.textContent = "Submitting…"; status.className = "status pending";
  try {
    const run = await createRun({ agent_id: agentInput.value.trim(), input: promptInput.value.trim() });
    status.textContent = `Run ${run.id} queued`;
    const completed = await watchRun(run.id);
    status.textContent = completed.status === "completed" ? "Completed" : "Failed";
    status.className = `status ${completed.status}`;
    result.textContent = completed.result || "The run did not return a result.";
  } catch (error) {
    status.textContent = error instanceof Error ? error.message : "Request failed";
    status.className = "status failed";
  } finally { submit.disabled = false; }
});