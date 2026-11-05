import{describe,expect,it}from"vitest";
import{directionFromDelta,directionFromFacing}from"./characterSprites";

describe("character sprite directions",()=>{
 it("maps movement into four 2.5D directions",()=>{expect(directionFromDelta(-1,1)).toBe("southwest");expect(directionFromDelta(1,1)).toBe("southeast");expect(directionFromDelta(-1,-1)).toBe("northwest");expect(directionFromDelta(1,-1)).toBe("northeast");});
 it("maps player facing into the same direction contract",()=>{expect(directionFromFacing(0)).toBe("southeast");expect(directionFromFacing(Math.PI)).toBe("northeast");});
});
