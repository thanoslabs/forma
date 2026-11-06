import{afterEach,describe,expect,it,vi}from"vitest";
import{agentStreamURL,listWorkspaceAgents,watchWorkspaceAgents}from"./agentApi";

afterEach(()=>vi.unstubAllGlobals());
describe("agent projections",()=>{
 it("loads workspace-scoped agents",async()=>{const fetch=vi.fn().mockResolvedValue(new Response(JSON.stringify([{connectionId:"connection-1",adapterId:"mock",id:"atlas",name:"Atlas",status:"idle",lastSeen:"2026-01-01T00:00:00Z"}]),{status:200}));vi.stubGlobal("fetch",fetch);await expect(listWorkspaceAgents("workspace / 1")).resolves.toHaveLength(1);expect(fetch.mock.calls[0][0]).toContain("workspace%20%2F%201/agents");});
 it("surfaces provider errors",async()=>{vi.stubGlobal("fetch",vi.fn().mockResolvedValue(new Response("provider unavailable",{status:502})));await expect(listWorkspaceAgents("workspace-1")).rejects.toThrow("provider unavailable");});
 it("builds secure stream URLs",()=>expect(agentStreamURL("workspace / 1",{protocol:"https:",host:"forma.example"} as Location)).toBe("wss://forma.example/api/v1/workspaces/workspace%20%2F%201/agents/ws"));
 it("projects streamed snapshots",()=>{class Socket{listeners=new Map<string,(event:any)=>void>();addEventListener(type:string,listener:(event:any)=>void){this.listeners.set(type,listener)}close(){}emit(type:string,event:any){this.listeners.get(type)?.(event)}}let socket:Socket|undefined;vi.stubGlobal("location",{protocol:"http:",host:"forma.test"});vi.stubGlobal("WebSocket",class extends Socket{constructor(){super();socket=this}});const receive=vi.fn(),stop=watchWorkspaceAgents("workspace-1",receive);socket!.emit("message",{data:JSON.stringify({type:"agents.snapshot",workspaceId:"workspace-1",agents:[{id:"atlas"}],observedAt:"now"})});expect(receive).toHaveBeenCalledWith([{id:"atlas"}]);stop();});
});
