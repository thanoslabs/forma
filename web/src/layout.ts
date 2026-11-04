import type { Position } from "./Office";
export type OfficeLayout={version:1; desks:Position[]};
export const defaultLayout:OfficeLayout={version:1,desks:[{x:220,y:205},{x:430,y:205},{x:640,y:205}]};
const KEY="forma:demo:layout:v1";
export function validateLayout(value:unknown):OfficeLayout {
 if(!value||typeof value!=="object")return defaultLayout;
 const candidate=value as Partial<OfficeLayout>;
 if(candidate.version!==1||!Array.isArray(candidate.desks)||candidate.desks.length!==3)return defaultLayout;
 const valid=candidate.desks.every(p=>p&&typeof p.x==="number"&&typeof p.y==="number"&&Number.isFinite(p.x)&&Number.isFinite(p.y)&&p.x>=110&&p.x<=750&&p.y>=130&&p.y<=290);
 return valid?{version:1,desks:candidate.desks.map(p=>({...p}))}:defaultLayout;
}
export function loadLayout():OfficeLayout {
 try{return validateLayout(JSON.parse(localStorage.getItem(KEY)??"null") as unknown);}catch{return defaultLayout;}
}
export function saveLayout(layout:OfficeLayout):void {
 try{localStorage.setItem(KEY,JSON.stringify(validateLayout(layout)));}catch{ /* storage may be unavailable */ }
}
export function resetLayout():void {try{localStorage.removeItem(KEY);}catch{ /* storage may be unavailable */ }}
