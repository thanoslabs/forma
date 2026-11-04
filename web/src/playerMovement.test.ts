import { describe, expect, it } from "vitest";
import { collidesWithDesk, inputVector, movePlayer, PLAYER_START } from "./playerMovement";

const desks=[{x:220,y:205}];
describe("player controls",()=>{
 it("starts within the office",()=>{expect(PLAYER_START).toEqual({x:125,y:375});});
 it("normalizes diagonals to avoid faster diagonal movement",()=>{const v=inputVector(new Set(["w","d"]));expect(Math.hypot(v.x,v.y)).toBeCloseTo(1);expect(v.x).toBeGreaterThan(0);expect(v.y).toBeLessThan(0);});
 it("supports arrow keys and opposite directions cancel",()=>{expect(inputVector(new Set(["arrowleft"]))).toEqual({x:-1,y:0});expect(inputVector(new Set(["a","d"]))).toEqual({x:0,y:0});});
 it("clamps movement at office boundaries",()=>{expect(movePlayer({x:72,y:95},-300,-300,[])).toEqual({x:72,y:95});});
 it("does not enter desk collision rectangles",()=>{expect(collidesWithDesk({x:220,y:205},desks)).toBe(true);expect(movePlayer({x:220,y:270},0,-90,desks)).toEqual({x:220,y:270});});
 it("slides along desk edges",()=>{expect(movePlayer({x:120,y:205},90,30,desks)).toEqual({x:120,y:235});});
});
