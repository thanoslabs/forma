import { describe,expect,it } from "vitest";
import { initialSimulation } from "./simulation";
import { validateSimulation } from "./simulationStorage";
describe("simulation persistence",()=>{
 it("rejects malformed snapshots",()=>{
  expect(validateSimulation({tick:1,agents:[]})).toEqual(initialSimulation());
 });
 it("restores valid snapshots paused",()=>{
  const state={...initialSimulation(),tick:5,running:true};
  const restored=validateSimulation(state);
  expect(restored.tick).toBe(5);
  expect(restored.running).toBe(false);
 });
});
