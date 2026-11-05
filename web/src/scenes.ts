import type{Workspace}from"./conversationApi";
import{tileToWorld,type GridPoint}from"./worldGrid";

export type SceneKind="office"|"street";
export type WorldScene={id:string;kind:SceneKind;name:string;floor?:number;workspaceId?:string;spawn:GridPoint;elevator?:GridPoint};
export const ELEVATOR_TILE:GridPoint={column:22,row:7};
export function workspaceScenes(workspaces:Workspace[]):WorldScene[]{return workspaces.map((workspace,index)=>({id:`workspace:${workspace.id}`,kind:"office",name:workspace.name,floor:index+1,workspaceId:workspace.id,spawn:{column:20,row:9},elevator:ELEVATOR_TILE}));}
export const sceneSpawn=(scene:WorldScene)=>tileToWorld(scene.spawn);
