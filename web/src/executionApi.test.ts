import{afterEach,describe,expect,it,vi}from"vitest";
import{listExecutions}from"./executionApi";
afterEach(()=>vi.unstubAllGlobals());
describe("execution API",()=>{it("loads workspace-scoped execution projections",async()=>{const mock=vi.fn().mockResolvedValue(new Response(JSON.stringify([{id:"execution-1"}]),{status:200}));vi.stubGlobal("fetch",mock);await expect(listExecutions("workspace / 1")).resolves.toEqual([{id:"execution-1"}]);expect(mock).toHaveBeenCalledWith("/api/v1/workspaces/workspace%20%2F%201/executions",{signal:undefined})});it("surfaces projection errors",async()=>{vi.stubGlobal("fetch",vi.fn().mockResolvedValue(new Response("bridge unavailable",{status:502})));await expect(listExecutions("workspace-1")).rejects.toThrow("bridge unavailable")})});
