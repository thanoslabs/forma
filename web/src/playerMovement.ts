import { useEffect, useRef, useState } from "react";
import { type Position } from "./Office";
import{deskBlockedTiles,moveOnGrid,snapToGrid,worldToTile}from"./worldGrid";

export const PLAYER_START: Position = snapToGrid({ x: 125, y: 375 });
export function collidesWithDesk(p:Position, desks:Position[]):boolean {
 return deskBlockedTiles(desks).has(`${worldToTile(p).column}:${worldToTile(p).row}`);
}
export function movePlayer(p:Position, dx:number, dy:number, desks:Position[]):Position {
 return moveOnGrid(p,{column:Math.sign(dx),row:Math.sign(dy)},deskBlockedTiles(desks));
}
export function inputVector(keys:ReadonlySet<string>):Position {
 const x=Number(keys.has("d")||keys.has("arrowright"))-Number(keys.has("a")||keys.has("arrowleft"));
 const y=Number(keys.has("s")||keys.has("arrowdown"))-Number(keys.has("w")||keys.has("arrowup"));
 const length=Math.hypot(x,y)||1;
 return {x:x/length,y:y/length};
}
export type PlayerState={position:Position;moving:boolean;facing:number};
export function usePlayerMovement(desks:Position[],enabled:boolean):PlayerState {
 const [state,setState]=useState<PlayerState>({position:PLAYER_START,moving:false,facing:0});
 const position=useRef<Position>(PLAYER_START),facing=useRef(0),desksRef=useRef(desks),lastMove=useRef(0),stopTimer=useRef<number|undefined>(undefined);
 desksRef.current=desks;
 useEffect(()=>{
  if(!enabled)return;
  const down=(event:KeyboardEvent)=>{
   const key=event.key.toLowerCase();
   if(!["w","a","s","d","arrowup","arrowdown","arrowleft","arrowright"].includes(key))return;
   const target=event.target;
   if(target instanceof HTMLElement && (target.isContentEditable||["INPUT","TEXTAREA","SELECT"].includes(target.tagName)))return;
   event.preventDefault();const now=performance.now();if(now-lastMove.current<105)return;lastMove.current=now;
   const vector=inputVector(new Set([key])),next=movePlayer(position.current,vector.x,vector.y,desksRef.current),actuallyMoved=next.x!==position.current.x||next.y!==position.current.y;
   if(!actuallyMoved)return;position.current=next;facing.current=Math.atan2(vector.x,vector.y);setState({position:next,moving:true,facing:facing.current});
   if(stopTimer.current)window.clearTimeout(stopTimer.current);stopTimer.current=window.setTimeout(()=>setState(value=>({...value,moving:false})),150);
  };
  const blur=()=>setState(s=>({...s,moving:false}));window.addEventListener("keydown",down);window.addEventListener("blur",blur);
  return()=>{if(stopTimer.current)window.clearTimeout(stopTimer.current);window.removeEventListener("keydown",down);window.removeEventListener("blur",blur);};
 },[enabled]);
 return state;
}
