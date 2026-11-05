import { useEffect, useRef } from "react";
import { Application, Container, Graphics, Text } from "pixi.js";
import { makeElevator,makeProjectBoard, makeWindows, makeWorkstations } from "./OfficeFixtures";
import {createCharacterSprite,directionFromDelta,directionFromFacing,preloadCharacterSprites,type CharacterDirection,type CharacterSprite} from "./characterSprites";
import type {PlayerState} from "./playerMovement";
import{snapToGrid}from"./worldGrid";
import{makeStreetScene}from"./StreetScene";

export type Position = { x: number; y: number };
export const desks: Position[] = [{ x: 220, y: 205 }, { x: 430, y: 205 }, { x: 640, y: 205 }];
export const bounds = { left: 22, right: 838, top: 26, bottom: 464 };
export function clampPosition(p: Position): Position {
 return { x: Math.max(bounds.left, Math.min(bounds.right, p.x)), y: Math.max(bounds.top, Math.min(bounds.bottom, p.y)) };
}
export function walkToward(current: Position, target: Position, distance: number): Position {
 const dx = target.x - current.x, dy = target.y - current.y, length = Math.hypot(dx, dy);
 if (length === 0 || length <= distance) return target;
 return { x: current.x + dx * distance / length, y: current.y + dy * distance / length };
}
export type OfficeAgent = { id: string; name: string; status: string; position: Position; color?: number; pose?:"stand"|"sit"|"work" };
export type OfficeProps = { agents?: OfficeAgent[]; playerState:PlayerState;sceneKind?:"office"|"street"; onSelect?: (id: string) => void; onDeselect?: () => void; onOpenProjects?: () => void;onOpenElevator?:()=>void;onEnterBuilding?:()=>void; onMove?: (id: string, position: Position) => void; selectedId?: string; deskPositions?: Position[] };
const fallback: OfficeAgent[] = [{ id: "demo-agent", name: "Demo Agent", status: "idle", position: { x: 430, y: 355 } }];
export function Office({ agents = fallback, playerState,sceneKind="office", onSelect, onDeselect, onOpenProjects,onOpenElevator,onEnterBuilding, onMove, selectedId, deskPositions = desks }: OfficeProps) {
 const host = useRef<HTMLDivElement>(null);
 const live = useRef({ agents, playerState, onSelect, onDeselect, onOpenProjects,onOpenElevator,onEnterBuilding, onMove, selectedId, deskPositions });
 live.current = { agents, playerState, onSelect, onDeselect, onOpenProjects,onOpenElevator,onEnterBuilding, onMove, selectedId, deskPositions };
 useEffect(() => {
  const element = host.current;
  if (!element) return;
  const app = new Application();
  let disposed = false;
  const actors = new Map<string, { node: Container; character:CharacterSprite; indicator:Graphics; indicatorText:Text; label: Text; position: Position; target: Position; source: Position; direction:CharacterDirection }>();
  let drag: { id: string; pointerId: number; start: Position; moved: boolean } | undefined;
  let observerCleanup: (() => void) | undefined;
  const init = async () => {
   await app.init({ width: 860, height: 490, background: "#c4e5ed", antialias: true, resolution: Math.min(window.devicePixelRatio || 1, 2), autoDensity: true });
   if (disposed) { app.destroy(true); return; }
   element.appendChild(app.canvas);
   // Keep world coordinates logical (860×490), scaling the single Pixi stage on resize.
   const resize = () => {
    const width = Math.max(1, element.clientWidth);
    const height = Math.max(1, element.clientHeight);
    app.renderer.resize(width, height);
    app.stage.scale.set(width / 860, height / 490);
   };
   const observer = new ResizeObserver(resize);
   observer.observe(element);
   observerCleanup = () => observer.disconnect();
   resize();
   let streetUpdate:((deltaMS:number)=>void)|undefined;
   if(sceneKind==="office"){
    const room = new Graphics().rect(0, 0, 860, 490).fill("#f5dfb6");room.eventMode="static";room.cursor="crosshair";room.on("pointertap",()=>live.current.onDeselect?.());app.stage.addChild(room);app.stage.addChild(makeWindows(),makeProjectBoard(()=>live.current.onOpenProjects?.()),makeElevator(()=>live.current.onOpenElevator?.()));
   }else{const street=await makeStreetScene(()=>live.current.onEnterBuilding?.());app.stage.addChild(street.layer);streetUpdate=street.update;}
   app.stage.eventMode = "static";
   app.stage.on("globalpointermove", e => {
    if (!drag) return;
   const point = snapToGrid(clampPosition(e.getLocalPosition(app.stage)));
    if (Math.hypot(point.x - drag.start.x, point.y - drag.start.y) > 5) drag.moved = true;
    if (drag.moved) {
     const actor = actors.get(drag.id);
     if (actor) { actor.target = point; actor.position = point; }
     live.current.onMove?.(drag.id, point);
    }
   });
   const finishDrag = () => { drag = undefined; };
   app.stage.on("pointerup", finishDrag);
   app.stage.on("pointerupoutside", finishDrag);
   await preloadCharacterSprites();
   const world=new Container();world.sortableChildren=true;app.stage.addChild(world);
   const playerCharacter=createCharacterSprite();const playerNode=new Container();playerNode.addChild(playerCharacter.sprite);world.addChild(playerNode);let renderedPlayer={...live.current.playerState.position};
   let deskKey = "";
   let furniture:Container[]=[];const drawFurniture = (positions: Position[]) => { furniture.forEach(child=>{world.removeChild(child);child.destroy({children:true});});furniture=makeWorkstations(positions);world.addChild(...furniture); };
   if(sceneKind==="office")drawFurniture(live.current.deskPositions);
   app.ticker.add(ticker => {
    streetUpdate?.(ticker.deltaMS);const key = JSON.stringify(live.current.deskPositions); if (sceneKind==="office"&&key !== deskKey) { deskKey = key; drawFurniture(live.current.deskPositions); }
    const incoming = live.current.agents;
    for (const a of incoming) {
     let actor = actors.get(a.id);
     if (!actor) {
      const node = new Container(),character=createCharacterSprite(),indicator=new Graphics(),indicatorText=new Text({text:"",style:{fontFamily:"sans-serif",fontSize:15,fontWeight:"bold",fill:0xffffff}});indicator.circle(0,0,15).fill(0x59677b).stroke({color:0xffffff,width:2});indicator.position.set(0,-88);indicatorText.anchor.set(.5);indicatorText.position.set(0,-89);node.addChild(character.sprite,indicator,indicatorText);
      node.eventMode = "static"; node.cursor = "pointer";
      node.on("pointerdown", e => {
       e.stopPropagation();
       drag = { id: a.id, pointerId: e.pointerId, start: { ...(actors.get(a.id)?.position ?? a.position) }, moved: false };
       live.current.onSelect?.(a.id);
      });
      node.on("pointertap", e => e.stopPropagation());
      const label = new Text({ text: "", style: { fontFamily: "sans-serif", fontSize: 13, fill: "#28384a" } });
      label.anchor.set(0.5);
      world.addChild(node);app.stage.addChild(label);
      actor = { node,character,indicator,indicatorText,label, position: { ...a.position }, target: { ...a.position }, source: { ...a.position },direction:"southeast" };
      actors.set(a.id, actor);
     }
     if (actor.source.x !== a.position.x || actor.source.y !== a.position.y) { actor.target = clampPosition(a.position); actor.source = { ...a.position }; }
     const before={...actor.position};actor.position = drag?.id === a.id && drag.moved ? actor.target : walkToward(actor.position, actor.target, Math.min(12, 2.5 * ticker.deltaTime));
     const dx=actor.position.x-before.x,dy=actor.position.y-before.y,moving=Math.hypot(dx,dy)>.01;if(moving)actor.direction=directionFromDelta(dx,dy,actor.direction);
     actor.node.position.set(actor.position.x, actor.position.y);
     actor.label.position.set(actor.position.x, actor.position.y + 47);
     actor.label.text = a.name + " · " + a.status;
     actor.node.alpha = live.current.selectedId === a.id ? 1 : 0.96;actor.node.zIndex=actor.position.y-(a.pose==="sit"||a.pose==="work"?50:0);
     actor.character.setPose(moving?"walk":a.pose==="work"?"work":a.pose==="sit"?"sit":"idle",actor.direction,app.ticker.lastTime/1000);
     const indicator=a.status==="working"?{text:"⚙",color:0x3b8f79}:a.status==="walking"?{text:"…",color:0x627fbd}:a.status==="blocked"?{text:"!",color:0xb2574f}:a.status==="completed"?{text:"✓",color:0x678d50}:undefined;actor.indicator.visible=actor.indicatorText.visible=!!indicator;if(indicator){actor.indicator.clear().circle(0,0,15).fill(indicator.color).stroke({color:0xffffff,width:2});actor.indicatorText.text=indicator.text;}
     actor.label.style.fill = a.status === "working" ? "#275c50" : a.status === "blocked" ? "#a04640" : "#28384a";
    }
    const ps=live.current.playerState;renderedPlayer=walkToward(renderedPlayer,ps.position,Math.min(14,4*ticker.deltaTime));const playerMoving=Math.hypot(renderedPlayer.x-ps.position.x,renderedPlayer.y-ps.position.y)>.2;playerNode.position.set(renderedPlayer.x,renderedPlayer.y);playerNode.zIndex=renderedPlayer.y-(ps.seated?50:0);playerCharacter.setPose(playerMoving?"walk":ps.seated?"sit":"idle",directionFromFacing(ps.facing),app.ticker.lastTime/1000);
    for (const [id, actor] of actors) if (!incoming.some(a => a.id === id)) { actor.node.destroy(); actor.label.destroy(); actors.delete(id); }
   });
  };
  void init();
  return () => { disposed = true; observerCleanup?.(); if (app.renderer) app.destroy(true); };
 }, [sceneKind]);
 return <div ref={host} className="office-canvas" aria-label={sceneKind==="street"?"Interactive illustrated street. Use the keyboard to explore and enter the office building.":"Interactive illustrated office. Select an agent and use the keyboard to move."} />;
}
