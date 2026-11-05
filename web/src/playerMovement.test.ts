import{describe,expect,it}from"vitest";
import{collidesWithDesk,inputVector,movePlayer,PLAYER_START}from"./playerMovement";
import{snapToGrid}from"./worldGrid";

const desks=[snapToGrid({x:220,y:205})];
describe("grid player controls",()=>{
 it("starts on the office grid",()=>{expect(PLAYER_START).toEqual(snapToGrid(PLAYER_START));});
 it("normalizes diagonals",()=>{const v=inputVector(new Set(["w","d"]));expect(Math.hypot(v.x,v.y)).toBeCloseTo(1);});
 it("supports arrow keys and opposite directions cancel",()=>{expect(inputVector(new Set(["arrowleft"]))).toEqual({x:-1,y:0});expect(inputVector(new Set(["a","d"]))).toEqual({x:0,y:0});});
 it("moves exactly one tile and stays within bounds",()=>{expect(movePlayer(PLAYER_START,1,0,[]).x-PLAYER_START.x).toBe(35);expect(movePlayer({x:17.5,y:17.5},-1,0,[])).toEqual({x:17.5,y:17.5});});
 it("does not enter workstation cells",()=>{expect(collidesWithDesk(desks[0],desks)).toBe(true);const below=snapToGrid({x:220,y:250});expect(movePlayer(below,0,-1,desks)).toEqual(below);});
});
