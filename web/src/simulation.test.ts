import { describe,expect,it } from "vitest";
import { assignTask, initialSimulation, tickSimulation } from "./simulation";
describe("simulation",()=>{
 it("starts with three idle agents and queued tasks",()=>{
  const state=initialSimulation();expect(state.agents).toHaveLength(3);expect(state.tasks.every(t=>t.status==="queued")).toBe(true);
 });
 it("assigns once and completes a task",()=>{
  let state=assignTask(initialSimulation(),"design","atlas");
  expect(state.tasks[0].status).toBe("active");
  expect(assignTask(state,"design","nova")).toBe(state);
  for(let i=0;i<100;i++)state=tickSimulation(state);
  expect(state.tasks.find(t=>t.id==="design")?.status).toBe("done");
 });
 it("does not progress while paused",()=>{
  const state={...initialSimulation(),running:false};expect(tickSimulation(state)).toBe(state);
 });
});
