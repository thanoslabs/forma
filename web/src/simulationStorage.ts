import { initialSimulation, type SimState } from "./simulation";
const KEY="forma:simulation:v1";
export function validateSimulation(value:unknown):SimState {
 const fresh=initialSimulation();
 if(!value||typeof value!=="object")return fresh;
 const state=value as Partial<SimState>;
 if(typeof state.tick!=="number"||!Number.isSafeInteger(state.tick)||state.tick<0||
  typeof state.nextEvent!=="number"||!Number.isSafeInteger(state.nextEvent)||state.nextEvent<1||
  !Array.isArray(state.agents)||state.agents.length!==3||
  !Array.isArray(state.tasks)||state.tasks.length>30||!Array.isArray(state.events)||state.events.length>60||
  typeof state.speed!=="number"||![0.25,0.5,1,2,4].includes(state.speed))return fresh;
 const known=new Set(fresh.agents.map(a=>a.id));
 if(!state.agents.every(a=>a&&known.has(a.id)&&["idle","walking","working","blocked","completed"].includes(a.status)&&
  typeof a.progress==="number"&&a.progress>=0&&a.progress<=100&&
  a.position&&a.target&&[a.position.x,a.position.y,a.target.x,a.target.y].every(Number.isFinite)))return fresh;
 if(!state.tasks.every(t=>t&&typeof t.id==="string"&&typeof t.title==="string"&&["queued","active","done"].includes(t.status)))return fresh;
 if(!state.events.every(e=>e&&typeof e.message==="string"&&typeof e.tick==="number"&&typeof e.id==="number"))return fresh;
 return {...state,agents:state.agents.map(a=>({...a,position:{...a.position},target:{...a.target}})),tasks:state.tasks.map(t=>({...t})),events:state.events.map(e=>({...e})),running:false} as SimState;
}
export function loadSimulation():SimState {
 try{return validateSimulation(JSON.parse(localStorage.getItem(KEY)??"null") as unknown);}catch{return initialSimulation();}
}
export function saveSimulation(state:SimState):void {
 try{localStorage.setItem(KEY,JSON.stringify({...state,running:false}));}catch{ /* storage unavailable */ }
}
export function clearSimulation():void {try{localStorage.removeItem(KEY);}catch{ /* storage unavailable */ }}
