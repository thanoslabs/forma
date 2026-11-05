import { useEffect, useRef, useState } from "react";
import { type Position } from "./Office";
import{deskBlockedTiles,moveOnGrid,snapToGrid,worldToTile}from"./worldGrid";
import{createWorkstations}from"./workstations";

export const PLAYER_START: Position = snapToGrid({ x: 125, y: 375 });
export const MOVEMENT_INTERVAL_MS=130;
const movementKeys=["w","a","s","d","arrowup","arrowdown","arrowleft","arrowright"] as const;
export function isMovementKey(key:string):boolean{return (movementKeys as readonly string[]).includes(key.toLowerCase());}
export function collidesWithDesk(p:Position, desks:Position[]):boolean {
 return deskBlockedTiles(desks).has(`${worldToTile(p).column}:${worldToTile(p).row}`);
}
export function movePlayer(p:Position, dx:number,dy:number,desks:Position[],environmentBlocked:ReadonlySet<string>=new Set()):Position {
 return moveOnGrid(p,{column:Math.sign(dx),row:Math.sign(dy)},new Set([...deskBlockedTiles(desks),...environmentBlocked]));
}
export function availableChair(position:Position,desks:Position[],occupiedSeats:ReadonlySet<string>){return createWorkstations(desks).find(item=>item.chair.x===position.x&&item.chair.y===position.y&&!occupiedSeats.has(item.id));}
export function inputVector(keys:ReadonlySet<string>):Position {
 const x=Number(keys.has("d")||keys.has("arrowright"))-Number(keys.has("a")||keys.has("arrowleft"));
 const y=Number(keys.has("s")||keys.has("arrowdown"))-Number(keys.has("w")||keys.has("arrowup"));
 const length=Math.hypot(x,y)||1;
 return {x:x/length,y:y/length};
}
export type PlayerState={position:Position;moving:boolean;facing:number;seated:boolean;seatId?:string};
export function usePlayerMovement(desks:Position[],enabled:boolean,occupiedSeats:ReadonlySet<string>=new Set(),sceneId="office",spawn=PLAYER_START,environmentBlocked:ReadonlySet<string>=new Set()):PlayerState {
 const [state,setState]=useState<PlayerState>({position:PLAYER_START,moving:false,facing:0,seated:false});
 const position=useRef<Position>(spawn),facing=useRef(0),seated=useRef(false),seatId=useRef<string|undefined>(undefined),desksRef=useRef(desks),occupiedRef=useRef(occupiedSeats),blockedRef=useRef(environmentBlocked),sceneRef=useRef(sceneId),scenePositions=useRef(new Map<string,Position>());
 desksRef.current=desks;
 occupiedRef.current=occupiedSeats;
 blockedRef.current=environmentBlocked;
 useEffect(()=>{if(sceneRef.current===sceneId)return;scenePositions.current.set(sceneRef.current,position.current);sceneRef.current=sceneId;position.current=scenePositions.current.get(sceneId)??spawn;seated.current=false;seatId.current=undefined;setState({position:position.current,moving:false,facing:facing.current,seated:false});},[sceneId,spawn.x,spawn.y]);
 useEffect(()=>{
  if(!enabled)return;
  const held=new Set<string>();let movementTimer:number|undefined;
  const stop=(updateState=true)=>{if(movementTimer!==undefined)window.clearInterval(movementTimer);movementTimer=undefined;held.clear();if(updateState)setState(value=>({...value,moving:false}));};
  const step=()=>{
   const vector=inputVector(held);if(vector.x===0&&vector.y===0){setState(value=>({...value,moving:false}));return;}
   if(seated.current){const station=createWorkstations(desksRef.current).find(item=>item.id===seatId.current);if(station){position.current=station.approach;seated.current=false;seatId.current=undefined;setState({position:station.approach,moving:true,facing:facing.current,seated:false});}return;}
   const next=movePlayer(position.current,vector.x,vector.y,desksRef.current,blockedRef.current),actuallyMoved=next.x!==position.current.x||next.y!==position.current.y;
   if(!actuallyMoved){setState(value=>({...value,moving:false}));return;}position.current=next;facing.current=Math.atan2(vector.x,vector.y);setState({position:next,moving:true,facing:facing.current,seated:false});
   const station=availableChair(next,desksRef.current,occupiedRef.current);if(station){seated.current=true;seatId.current=station.id;facing.current=Math.PI;held.clear();if(movementTimer!==undefined)window.clearInterval(movementTimer);movementTimer=undefined;setState({position:station.chair,moving:false,facing:facing.current,seated:true,seatId:station.id});}
  };
  const down=(event:KeyboardEvent)=>{
   const key=event.key.toLowerCase();
   if(!isMovementKey(key))return;
   const target=event.target;
   if(target instanceof HTMLElement && (target.isContentEditable||["INPUT","TEXTAREA","SELECT"].includes(target.tagName)))return;
   event.preventDefault();if(held.has(key))return;const wasIdle=held.size===0;held.add(key);if(wasIdle){step();movementTimer=window.setInterval(step,MOVEMENT_INTERVAL_MS);}
  };
  const up=(event:KeyboardEvent)=>{const key=event.key.toLowerCase();if(!isMovementKey(key))return;held.delete(key);if(held.size===0)stop();};
  const blur=()=>stop();window.addEventListener("keydown",down);window.addEventListener("keyup",up);window.addEventListener("blur",blur);
  return()=>{stop(false);window.removeEventListener("keydown",down);window.removeEventListener("keyup",up);window.removeEventListener("blur",blur);};
 },[enabled,sceneId]);
 return state;
}
