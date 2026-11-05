import type{Workspace}from"./conversationApi";
import{tileToWorld,type GridPoint}from"./worldGrid";

export type SceneKind="office"|"street";
export type WorldScene={id:string;kind:SceneKind;name:string;floor?:number;workspaceId?:string;spawn:GridPoint;elevator?:GridPoint};
export const ELEVATOR_TILE:GridPoint={column:22,row:7};
export function workspaceScenes(workspaces:Workspace[]):WorldScene[]{return workspaces.map((workspace,index)=>({id:`workspace:${workspace.id}`,kind:"office",name:workspace.name,floor:index+1,workspaceId:workspace.id,spawn:{column:20,row:9},elevator:ELEVATOR_TILE}));}
export const STREET_SCENE:WorldScene={id:"street",kind:"street",name:"Forma Street",spawn:{column:12,row:12}};
export const DEMO_WORKSPACES:Workspace[]=[{id:"demo-studio",name:"Agent Studio"},{id:"demo-operations",name:"Operations"},{id:"demo-lab",name:"Research Lab"}];
export const sceneSpawn=(scene:WorldScene)=>tileToWorld(scene.spawn);
export function streetBlockedTiles():Set<string>{const blocked=new Set<string>();for(let column=0;column<24;column++)for(let row=0;row<14;row++){const road=row>=7&&row<=11,crosswalk=column>=10&&column<=13;if(road&&!crosswalk)blocked.add(`${column}:${row}`);if(row<=4&&(column<=9||column>=18))blocked.add(`${column}:${row}`);}return blocked;}
