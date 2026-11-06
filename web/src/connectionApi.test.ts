import{afterEach,describe,expect,it,vi}from"vitest";
import{createConnection,listConnections,testConnection}from"./connectionApi";
afterEach(()=>vi.unstubAllGlobals());
const response=(body:unknown,status=200)=>new Response(typeof body==="string"?body:JSON.stringify(body),{status,headers:{"Content-Type":"application/json"}});
describe("connection API",()=>{
 it("lists workspace connections",async()=>{const fetch=vi.fn().mockResolvedValue(response([{id:"c1"}]));vi.stubGlobal("fetch",fetch);await expect(listConnections("workspace / 1")).resolves.toHaveLength(1);expect(fetch.mock.calls[0][0]).toContain("workspace%20%2F%201/connections");});
 it("creates a write-only credential connection",async()=>{const fetch=vi.fn().mockResolvedValue(response({id:"c1"},201));vi.stubGlobal("fetch",fetch);await createConnection("w1",{adapterId:"paseo",providerType:"agent",displayName:"Local Paseo",endpointUrl:"ws://localhost:7777",credential:"secret"});const init=fetch.mock.calls[0][1] as RequestInit;expect(init.method).toBe("POST");expect(JSON.parse(String(init.body))).toMatchObject({adapterId:"paseo",credential:"secret"});});
 it("tests discovery for one connection",async()=>{const fetch=vi.fn().mockResolvedValue(response({connectionId:"c1",status:"connected",agentCount:3,capabilities:["observe"]}));vi.stubGlobal("fetch",fetch);await expect(testConnection("w1","c1")).resolves.toMatchObject({agentCount:3});expect(fetch.mock.calls[0][0]).toContain("/connections/c1/test");});
});
