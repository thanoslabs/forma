import { Assets, Rectangle, Sprite, Texture } from "pixi.js";
import{publicAsset}from"./publicAsset";

export type CharacterDirection="southwest"|"southeast"|"northwest"|"northeast";
export type CharacterAnimation="idle"|"walk"|"sit"|"work";
export type CharacterKind="agent"|"player-tech"|"player-inventor";
export type CharacterSprite={sprite:Sprite;setPose:(animation:CharacterAnimation,direction:CharacterDirection,time:number)=>void};

const FRAME_WIDTH=300,FRAME_HEIGHT=300;
const columns:Record<CharacterDirection,number>={southwest:0,southeast:1,northwest:2,northeast:3};
const rows={idle:0,walkA:1,walkB:2,sit:3,work:4} as const;
const paths:Record<CharacterKind,string>={agent:"assets/characters/forma-agent/atlas.png","player-tech":"assets/characters/forma-player-tech/atlas.png","player-inventor":"assets/characters/forma-player-inventor/atlas.png"};
const texturePromises=new Map<CharacterKind,Promise<Texture[][]>>(),loadedTextures=new Map<CharacterKind,Texture[][]>();

function loadTextures(kind:CharacterKind):Promise<Texture[][]>{
 let promise=texturePromises.get(kind);if(!promise){promise=Assets.load(publicAsset(paths[kind])).then(texture=>{texture.source.scaleMode="nearest";const textures=Array.from({length:5},(_,row)=>Array.from({length:4},(_,column)=>new Texture({source:texture.source,frame:new Rectangle(column*FRAME_WIDTH,row*FRAME_HEIGHT,FRAME_WIDTH,FRAME_HEIGHT)})));loadedTextures.set(kind,textures);return textures;});texturePromises.set(kind,promise);}return promise;
}

export async function preloadCharacterSprites():Promise<void>{await Promise.all((Object.keys(paths) as CharacterKind[]).map(loadTextures));}
export function createCharacterSprite(kind:CharacterKind="agent"):CharacterSprite{
 const textures=loadedTextures.get(kind);if(!textures)throw new Error("character sprites must be preloaded");
 const sprite=new Sprite(textures[0][1]);
 sprite.anchor.set(.5,.88);sprite.scale.set(.42);
 return{sprite,setPose:(animation,direction,time)=>{
  const row=animation==="walk"?(Math.floor(time*6)%2?rows.walkA:rows.walkB):rows[animation];
  sprite.texture=textures[row][columns[direction]];
 }};
}

export function directionFromDelta(dx:number,dy:number,fallback:CharacterDirection="southeast"):CharacterDirection{
 if(Math.abs(dx)<.01&&Math.abs(dy)<.01)return fallback;
 return dy<0?(dx<0?"northwest":"northeast"):(dx<0?"southwest":"southeast");
}

export function directionFromFacing(facing:number):CharacterDirection{
 const x=Math.sin(facing),y=Math.cos(facing);return directionFromDelta(x,y);
}
