import{describe,expect,it}from"vitest";
import{workspaceScenes}from"./scenes";
describe("workspace scenes",()=>{it("maps each workspace to a stable floor",()=>{const scenes=workspaceScenes([{id:"a",name:"Alpha"},{id:"b",name:"Beta"}]);expect(scenes.map(scene=>[scene.id,scene.floor,scene.workspaceId])).toEqual([["workspace:a",1,"a"],["workspace:b",2,"b"]]);});});
