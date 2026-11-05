import type{Position}from"./Office";
import{tileToWorld,worldToTile,type GridPoint}from"./worldGrid";

export type Workstation={id:string;desk:Position;anchor:GridPoint;chair:Position;approach:Position;seatDirection:"northwest"|"northeast";blocked:GridPoint[]};

export function workstationFromDesk(desk:Position,index:number):Workstation{
 const anchor=worldToTile(desk),chairTile={column:anchor.column,row:anchor.row+2},approachTile={column:anchor.column,row:anchor.row+3};
 const blocked:GridPoint[]=[];for(let column=anchor.column-2;column<=anchor.column+2;column++)for(let row=anchor.row-1;row<=anchor.row+1;row++)blocked.push({column,row});
 return{id:`workstation-${index+1}`,desk:tileToWorld(anchor),anchor,chair:tileToWorld(chairTile),approach:tileToWorld(approachTile),seatDirection:index%2?"northwest":"northeast",blocked};
}
export function createWorkstations(desks:Position[]):Workstation[]{return desks.map(workstationFromDesk);}
