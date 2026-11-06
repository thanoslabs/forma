import { Assets, Rectangle, Sprite, Texture } from "pixi.js";
import{publicAsset}from"./publicAsset";

export type CharacterDirection="southwest"|"southeast"|"northwest"|"northeast";
export type CharacterAnimation="idle"|"walk"|"sit"|"work";
export type CharacterKind="agent"|"player";
export type CharacterSprite={sprite:Sprite;setPose:(animation:CharacterAnimation,direction:CharacterDirection,time:number)=>void};

const FRAME_SIZE=32,COLUMNS=4;
const paths:Record<CharacterKind,string>={agent:"assets/characters/forma-agent/atlas.png",player:"assets/characters/forma-player/atlas.png"};
const texturePromises=new Map<CharacterKind,Promise<Texture[][]>>(),loadedTextures=new Map<CharacterKind,Texture[][]>();

function loadTextures(kind:CharacterKind):Promise<Texture[][]>{
 let promise=texturePromises.get(kind);if(!promise){promise=Assets.load(publicAsset(paths[kind])).then(texture=>{texture.source.scaleMode="nearest";const textures=Array.from({length:2},(_,row)=>Array.from({length:COLUMNS},(_,column)=>new Texture({source:texture.source,frame:new Rectangle(column*FRAME_SIZE,row*FRAME_SIZE,FRAME_SIZE,FRAME_SIZE)})));loadedTextures.set(kind,textures);return textures;});texturePromises.set(kind,promise);}return promise;
}

export async function preloadCharacterSprites():Promise<void>{await Promise.all((Object.keys(paths) as CharacterKind[]).map(loadTextures));}
export function createCharacterSprite(kind:CharacterKind="agent"):CharacterSprite{
 const textures=loadedTextures.get(kind);if(!textures)throw new Error("character sprites must be preloaded");
 const sprite=new Sprite(textures[0][0]);sprite.anchor.set(.5,.82);sprite.scale.set(kind==="player"?5:6);
 return{sprite,setPose:(animation,direction,time)=>{const moving=animation==="walk",working=animation==="work",row=kind==="player"&&moving?1:0,frame=Math.floor(time*(working?4:moving?10:3))%COLUMNS;sprite.texture=textures[row][frame];const west=direction==="southwest"||direction==="northwest";sprite.scale.x=Math.abs(sprite.scale.x)*(west?-1:1);sprite.y=kind==="agent"?Math.round(Math.sin(time*4)*2):0;}};
}

export function directionFromDelta(dx:number,dy:number,fallback:CharacterDirection="southeast"):CharacterDirection{
 if(Math.abs(dx)<.01&&Math.abs(dy)<.01)return fallback;
 return dy<0?(dx<0?"northwest":"northeast"):(dx<0?"southwest":"southeast");
}

export function directionFromFacing(facing:number):CharacterDirection{
 const x=Math.sin(facing),y=Math.cos(facing);return directionFromDelta(x,y);
}
