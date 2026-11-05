import type{Position}from"./Office";

export type GridPoint={column:number;row:number};
export const TILE_SIZE=35,GRID_COLUMNS=24,GRID_ROWS=14;

export function tileToWorld(tile:GridPoint):Position{return{x:tile.column*TILE_SIZE+TILE_SIZE/2,y:tile.row*TILE_SIZE+TILE_SIZE/2};}
export function worldToTile(position:Position):GridPoint{return{column:Math.max(0,Math.min(GRID_COLUMNS-1,Math.round((position.x-TILE_SIZE/2)/TILE_SIZE))),row:Math.max(0,Math.min(GRID_ROWS-1,Math.round((position.y-TILE_SIZE/2)/TILE_SIZE)))};}
export function snapToGrid(position:Position):Position{return tileToWorld(worldToTile(position));}
export const tileKey=(tile:GridPoint)=>tile.column+":"+tile.row;

export function deskBlockedTiles(desks:Position[]):Set<string>{
 const blocked=new Set<string>();
 for(const position of desks){const anchor=worldToTile(position);for(let column=anchor.column-2;column<=anchor.column+2;column++)for(let row=anchor.row-1;row<=anchor.row+1;row++)if(inBounds({column,row}))blocked.add(tileKey({column,row}));}
 return blocked;
}
export function inBounds(tile:GridPoint):boolean{return tile.column>=0&&tile.column<GRID_COLUMNS&&tile.row>=0&&tile.row<GRID_ROWS;}

const neighbors=(tile:GridPoint):GridPoint[]=>[{column:tile.column+1,row:tile.row},{column:tile.column-1,row:tile.row},{column:tile.column,row:tile.row+1},{column:tile.column,row:tile.row-1}].filter(inBounds);
const distance=(a:GridPoint,b:GridPoint)=>Math.abs(a.column-b.column)+Math.abs(a.row-b.row);
export function findPath(start:GridPoint,goal:GridPoint,blocked:Set<string>):GridPoint[]{
 if(tileKey(start)===tileKey(goal))return[start];if(blocked.has(tileKey(goal)))return[];
 const open=[start],came=new Map<string,GridPoint>(),cost=new Map([[tileKey(start),0]]);
 while(open.length){open.sort((a,b)=>(cost.get(tileKey(a))??Infinity)+distance(a,goal)-(cost.get(tileKey(b))??Infinity)-distance(b,goal));const current=open.shift()!;if(tileKey(current)===tileKey(goal)){const path=[current];let cursor=current;while(came.has(tileKey(cursor))){cursor=came.get(tileKey(cursor))!;path.unshift(cursor);}return path;}
  for(const next of neighbors(current)){if(blocked.has(tileKey(next)))continue;const nextCost=(cost.get(tileKey(current))??0)+1;if(nextCost<(cost.get(tileKey(next))??Infinity)){cost.set(tileKey(next),nextCost);came.set(tileKey(next),current);if(!open.some(item=>tileKey(item)===tileKey(next)))open.push(next);}}
 }
 return[];
}

export function moveOnGrid(position:Position,delta:GridPoint,blocked:Set<string>):Position{
 const current=worldToTile(position),next={column:current.column+delta.column,row:current.row+delta.row};return inBounds(next)&&!blocked.has(tileKey(next))?tileToWorld(next):tileToWorld(current);
}
