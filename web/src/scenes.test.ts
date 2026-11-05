import{describe,expect,it}from"vitest";
import{DEMO_WORKSPACES,streetBlockedTiles,workspaceScenes}from"./scenes";
describe("workspace scenes",()=>{it("maps each workspace to a stable floor",()=>{const scenes=workspaceScenes([{id:"a",name:"Alpha"},{id:"b",name:"Beta"}]);expect(scenes.map(scene=>[scene.id,scene.floor,scene.workspaceId])).toEqual([["workspace:a",1,"a"],["workspace:b",2,"b"]]);});});
describe("Pages demo",()=>{it("provides several floors without a backend",()=>{expect(workspaceScenes(DEMO_WORKSPACES).map(scene=>scene.name)).toEqual(["Agent Studio","Operations","Research Lab"]);});});
describe("street scene",()=>{it("keeps traffic lanes blocked except at the crosswalk",()=>{const blocked=streetBlockedTiles();expect(blocked.has("4:9")).toBe(true);expect(blocked.has("11:9")).toBe(false);expect(blocked.has("4:12")).toBe(false);});});
