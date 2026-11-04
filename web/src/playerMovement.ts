import { useEffect, useRef, useState } from "react";
import { bounds, type Position } from "./Office";

export const PLAYER_START: Position = { x: 125, y: 375 };
export const PLAYER_SPEED = 155; // office pixels per second
const RADIUS = 18;
export function collidesWithDesk(p:Position, desks:Position[]):boolean {
 return desks.some(d => p.x > d.x-65-RADIUS && p.x < d.x+65+RADIUS && p.y > d.y-35-RADIUS && p.y < d.y+25+RADIUS);
}
export function movePlayer(p:Position, dx:number, dy:number, desks:Position[]):Position {
 const next={x:Math.max(bounds.left,Math.min(bounds.right,p.x+dx)),y:Math.max(bounds.top,Math.min(bounds.bottom,p.y+dy))};
 // Resolve axes independently to permit sliding along desk edges.
 const x={x:next.x,y:p.y};
 const horizontal=collidesWithDesk(x,desks)?p:x;
 const y={x:horizontal.x,y:next.y};
 return collidesWithDesk(y,desks)?horizontal:y;
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
 const keys=useRef(new Set<string>()),position=useRef<Position>(PLAYER_START),facing=useRef(0),desksRef=useRef(desks);
 desksRef.current=desks;
 useEffect(()=>{
  if(!enabled){keys.current.clear();return;}
  const down=(event:KeyboardEvent)=>{
   const key=event.key.toLowerCase();
   if(!["w","a","s","d","arrowup","arrowdown","arrowleft","arrowright"].includes(key))return;
   const target=event.target;
   if(target instanceof HTMLElement && (target.isContentEditable||["INPUT","TEXTAREA","SELECT"].includes(target.tagName)))return;
   event.preventDefault();keys.current.add(key);
  };
  const up=(event:KeyboardEvent)=>{keys.current.delete(event.key.toLowerCase());};
  const blur=()=>{keys.current.clear();setState(s=>({...s,moving:false}));};
  window.addEventListener("keydown",down);window.addEventListener("keyup",up);window.addEventListener("blur",blur);
  let frame=0,last=0;
  const tick=(now:number)=>{
   if(!last)last=now;
   const dt=Math.min((now-last)/1000,0.05);last=now;
   const vector=inputVector(keys.current),moving=vector.x!==0||vector.y!==0;
   if(moving){
    const next=movePlayer(position.current,vector.x*PLAYER_SPEED*dt,vector.y*PLAYER_SPEED*dt,desksRef.current);
    const actuallyMoved=next.x!==position.current.x||next.y!==position.current.y;
    position.current=next;
    if(actuallyMoved)facing.current=Math.atan2(vector.x,vector.y);
    setState({position:next,moving:actuallyMoved,facing:facing.current});
   }else setState(s=>s.moving?{...s,moving:false}:s);
   frame=requestAnimationFrame(tick);
  };
  frame=requestAnimationFrame(tick);
  return()=>{cancelAnimationFrame(frame);keys.current.clear();window.removeEventListener("keydown",down);window.removeEventListener("keyup",up);window.removeEventListener("blur",blur);};
 },[enabled]);
 return state;
}
