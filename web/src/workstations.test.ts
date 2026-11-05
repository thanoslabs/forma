import{describe,expect,it}from"vitest";
import{snapToGrid,worldToTile}from"./worldGrid";
import{workstationFromDesk}from"./workstations";

describe("workstation geometry",()=>{
 it("defines grid-aligned desk, chair, approach and footprint",()=>{const station=workstationFromDesk(snapToGrid({x:220,y:205}),0);expect(worldToTile(station.chair).row).toBe(station.anchor.row+2);expect(worldToTile(station.approach).row).toBe(station.anchor.row+3);expect(station.blocked).toHaveLength(15);});
});
