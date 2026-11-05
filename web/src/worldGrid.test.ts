import{describe,expect,it}from"vitest";
import{findPath,moveOnGrid,snapToGrid,tileKey,tileToWorld,worldToTile}from"./worldGrid";

describe("world grid",()=>{
 it("round trips and snaps world positions",()=>{const tile={column:7,row:9};expect(worldToTile(tileToWorld(tile))).toEqual(tile);expect(snapToGrid({x:251,y:338})).toEqual(tileToWorld(tile));});
 it("routes around blocked cells",()=>{const blocked=new Set([tileKey({column:1,row:0})]);const path=findPath({column:0,row:0},{column:2,row:0},blocked);expect(path[0]).toEqual({column:0,row:0});expect(path.at(-1)).toEqual({column:2,row:0});expect(path).not.toContainEqual({column:1,row:0});});
 it("refuses blocked movement",()=>{const start=tileToWorld({column:0,row:0});expect(moveOnGrid(start,{column:1,row:0},new Set(["1:0"]))).toEqual(start);});
});
