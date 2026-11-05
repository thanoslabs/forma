import { useEffect, useRef } from "react";
import { Application, Container, Graphics, Text } from "pixi.js";
import { makeProjectBoard, makeWindows, makeWorkstations } from "./OfficeFixtures";
import {createCharacterSprite,directionFromDelta,directionFromFacing,preloadCharacterSprites,type CharacterDirection,type CharacterSprite} from "./characterSprites";
import type {PlayerState} from "./playerMovement";

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
export type OfficeAgent = { id: string; name: string; status: string; position: Position; color?: number };
export type OfficeProps = { agents?: OfficeAgent[]; playerState:PlayerState; onSelect?: (id: string) => void; onDeselect?: () => void; onOpenProjects?: () => void; onMove?: (id: string, position: Position) => void; selectedId?: string; deskPositions?: Position[] };
const fallback: OfficeAgent[] = [{ id: "demo-agent", name: "Demo Agent", status: "idle", position: { x: 430, y: 355 } }];
export function Office({ agents = fallback, playerState, onSelect, onDeselect, onOpenProjects, onMove, selectedId, deskPositions = desks }: OfficeProps) {
 const host = useRef<HTMLDivElement>(null);
 const live = useRef({ agents, playerState, onSelect, onDeselect, onOpenProjects, onMove, selectedId, deskPositions });
 live.current = { agents, playerState, onSelect, onDeselect, onOpenProjects, onMove, selectedId, deskPositions };
 useEffect(() => {
  const element = host.current;
  if (!element) return;
  const app = new Application();
  let disposed = false;
  const actors = new Map<string, { node: Container; character:CharacterSprite; label: Text; position: Position; target: Position; source: Position; direction:CharacterDirection }>();
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
   const room = new Graphics().rect(0, 0, 860, 490).fill("#f5dfb6");
   room.eventMode = "static";
   room.cursor = "crosshair";
   room.on("pointertap", () => live.current.onDeselect?.());
   app.stage.addChild(room);
   app.stage.addChild(makeWindows());
   app.stage.addChild(makeProjectBoard(() => live.current.onOpenProjects?.()));
   app.stage.eventMode = "static";
   app.stage.on("globalpointermove", e => {
    if (!drag) return;
    const point = clampPosition(e.getLocalPosition(app.stage));
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
   const furniture = new Container(); app.stage.addChild(furniture);
   const characters=new Container();characters.sortableChildren=true;app.stage.addChild(characters);
   const playerCharacter=createCharacterSprite();const playerNode=new Container();playerNode.addChild(playerCharacter.sprite);characters.addChild(playerNode);
   let deskKey = "";
   const drawFurniture = (positions: Position[]) => { furniture.removeChildren().forEach(child => child.destroy({children:true})); furniture.addChild(makeWorkstations(positions)); };
   drawFurniture(live.current.deskPositions);
   app.ticker.add(ticker => {
    const key = JSON.stringify(live.current.deskPositions); if (key !== deskKey) { deskKey = key; drawFurniture(live.current.deskPositions); }
    const incoming = live.current.agents;
    for (const a of incoming) {
     let actor = actors.get(a.id);
     if (!actor) {
      const node = new Container(),character=createCharacterSprite();node.addChild(character.sprite);
      node.eventMode = "static"; node.cursor = "pointer";
      node.on("pointerdown", e => {
       e.stopPropagation();
       drag = { id: a.id, pointerId: e.pointerId, start: { ...(actors.get(a.id)?.position ?? a.position) }, moved: false };
       live.current.onSelect?.(a.id);
      });
      node.on("pointertap", e => e.stopPropagation());
      const label = new Text({ text: "", style: { fontFamily: "sans-serif", fontSize: 13, fill: "#28384a" } });
      label.anchor.set(0.5);
      characters.addChild(node);app.stage.addChild(label);
      actor = { node,character,label, position: { ...a.position }, target: { ...a.position }, source: { ...a.position },direction:"southeast" };
      actors.set(a.id, actor);
     }
     if (actor.source.x !== a.position.x || actor.source.y !== a.position.y) { actor.target = clampPosition(a.position); actor.source = { ...a.position }; }
     const before={...actor.position};actor.position = drag?.id === a.id && drag.moved ? actor.target : walkToward(actor.position, actor.target, Math.min(12, 2.5 * ticker.deltaTime));
     const dx=actor.position.x-before.x,dy=actor.position.y-before.y,moving=Math.hypot(dx,dy)>.01;if(moving)actor.direction=directionFromDelta(dx,dy,actor.direction);
     actor.node.position.set(actor.position.x, actor.position.y);
     actor.label.position.set(actor.position.x, actor.position.y + 47);
     actor.label.text = a.name + " · " + a.status;
     actor.node.alpha = live.current.selectedId === a.id ? 1 : 0.96;actor.node.zIndex=actor.position.y;
     actor.character.setPose(moving?"walk":a.status==="working"?"work":"idle",actor.direction,app.ticker.lastTime/1000);
     actor.label.style.fill = a.status === "working" ? "#275c50" : a.status === "blocked" ? "#a04640" : "#28384a";
    }
    const ps=live.current.playerState;playerNode.position.set(ps.position.x,ps.position.y);playerNode.zIndex=ps.position.y;playerCharacter.setPose(ps.moving?"walk":"idle",directionFromFacing(ps.facing),app.ticker.lastTime/1000);
    for (const [id, actor] of actors) if (!incoming.some(a => a.id === id)) { actor.node.destroy(); actor.label.destroy(); actors.delete(id); }
   });
  };
  void init();
  return () => { disposed = true; observerCleanup?.(); if (app.renderer) app.destroy(true); };
 }, []);
 return <div ref={host} className="office-canvas" aria-label="Interactive illustrated office. Select an agent and click the floor to move." />;
}
