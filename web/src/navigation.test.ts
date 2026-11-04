import { describe, expect, it } from "vitest";
import { avoidFurniture, clampPosition, walkToward } from "./Office";
describe("office navigation",()=>{
 it("clamps to room bounds",()=>expect(clampPosition({x:-100,y:900})).toEqual({x:72,y:423}));
 it("keeps floor clicks away from desks",()=>expect(avoidFurniture({x:220,y:210},[{x:220,y:205}]).y).toBe(268));
 it("walks without overshooting",()=>expect(walkToward({x:0,y:0},{x:3,y:4},10)).toEqual({x:3,y:4}));
});
