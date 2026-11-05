import{afterEach,describe,expect,it,vi}from"vitest";
import{listWorkspaceAgents}from"./agentApi";

afterEach(()=>vi.unstubAllGlobals());
describe("agent projections",()=>{
 it("loads workspace-scoped agents",async()=>{const fetch=vi.fn().mockResolvedValue(new Response(JSON.stringify([{connectionId:"connection-1",adapterId:"mock",id:"atlas",name:"Atlas",status:"idle",lastSeen:"2026-01-01T00:00:00Z"}]),{status:200}));vi.stubGlobal("fetch",fetch);await expect(listWorkspaceAgents("workspace / 1")).resolves.toHaveLength(1);expect(fetch.mock.calls[0][0]).toContain("workspace%20%2F%201/agents");});
 it("surfaces provider errors",async()=>{vi.stubGlobal("fetch",vi.fn().mockResolvedValue(new Response("provider unavailable",{status:502})));await expect(listWorkspaceAgents("workspace-1")).rejects.toThrow("provider unavailable");});
});
