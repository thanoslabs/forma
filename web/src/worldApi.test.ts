import{afterEach,describe,expect,it,vi}from"vitest";
import{getWorld,putWorld}from"./worldApi";

afterEach(()=>vi.unstubAllGlobals());
describe("workspace worlds",()=>{it("loads a validated layout",async()=>{vi.stubGlobal("fetch",vi.fn().mockResolvedValue(new Response(JSON.stringify({version:1,desks:[{x:220,y:205},{x:430,y:205},{x:640,y:205}]}),{status:200})));await expect(getWorld("workspace 1")).resolves.toMatchObject({version:1});});it("saves with PUT",async()=>{const fetch=vi.fn().mockResolvedValue(new Response(JSON.stringify({version:1,desks:[{x:220,y:205},{x:430,y:205},{x:640,y:205}]}),{status:200}));vi.stubGlobal("fetch",fetch);await putWorld("workspace-1",{version:1,desks:[{x:220,y:205},{x:430,y:205},{x:640,y:205}]});expect(fetch.mock.calls[0][1]).toMatchObject({method:"PUT"});});});
