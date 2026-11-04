import type { OfficeAgent, Position } from "./Office";

export type SimStatus = "idle" | "walking" | "working" | "blocked" | "completed";
export type SimAgent = OfficeAgent & { status: SimStatus; taskId?: string; progress: number; target: Position };
export type SimTask = { id: string; title: string; status: "queued" | "active" | "done"; agentId?: string };
export type SimEvent = { id: number; tick: number; message: string; agentId?: string };
export type SimState = { tick: number; agents: SimAgent[]; tasks: SimTask[]; events: SimEvent[]; running: boolean; speed: number; nextEvent: number };
const initialAgents: SimAgent[] = [
 { id:"atlas",name:"Atlas",status:"idle",position:{x:180,y:350},target:{x:180,y:350},progress:0,color:0x7e70cf },
 { id:"nova",name:"Nova",status:"idle",position:{x:430,y:350},target:{x:430,y:350},progress:0,color:0xe58d9a },
 { id:"pixel",name:"Pixel",status:"idle",position:{x:680,y:350},target:{x:680,y:350},progress:0,color:0x6abbb1 }
];
const initialTasks: SimTask[] = [
 {id:"design",title:"Design the welcome screen",status:"queued"},
 {id:"tests",title:"Write regression tests",status:"queued"},
 {id:"docs",title:"Draft release notes",status:"queued"}
];
export function initialSimulation(): SimState {
 return {tick:0,agents:initialAgents.map(a=>({...a,position:{...a.position},target:{...a.target}})),tasks:initialTasks.map(t=>({...t})),events:[],running:true,speed:1,nextEvent:1};
}
export function addEvent(state: SimState,message:string,agentId?:string): SimState {
 return {...state,nextEvent:state.nextEvent+1,events:[{id:state.nextEvent,tick:state.tick,message,agentId},...state.events].slice(0,60)};
}
export function assignTask(state: SimState,taskId:string,agentId:string): SimState {
 const task=state.tasks.find(t=>t.id===taskId),agent=state.agents.find(a=>a.id===agentId);
 if (!task || !agent || task.status!=="queued" || (agent.status!=="idle"&&agent.status!=="completed")) return state;
 const index=state.agents.findIndex(a=>a.id===agentId);
 const desk={x:[220,430,640][index]??430,y:285};
 const next={...state,tasks:state.tasks.map(t=>t.id===taskId?{...t,status:"active" as const,agentId}:t),agents:state.agents.map(a=>a.id===agentId?{...a,status:"walking" as const,taskId,progress:0,target:desk}:a)};
 return addEvent(next,agent.name+" accepted "+task.title,agentId);
}
export function tickSimulation(state:SimState):SimState {
 if(!state.running)return state;
 let next:SimState={...state,tick:state.tick+1,agents:state.agents.map(a=>({...a,position:{...a.position},target:{...a.target}})),tasks:state.tasks.map(t=>({...t}))};
 for(const agent of next.agents){
  if(agent.status==="walking"){
   const dx=agent.target.x-agent.position.x,dy=agent.target.y-agent.position.y,dist=Math.hypot(dx,dy);
   if(dist<=18){agent.position={...agent.target};agent.status="working";next=addEvent(next,agent.name+" started working",agent.id);}
   else{agent.position={x:agent.position.x+dx/dist*18,y:agent.position.y+dy/dist*18};}
  }else if(agent.status==="working"){
   agent.progress=Math.min(100,agent.progress+4);
   if(agent.progress>=100){
    agent.status="completed";
    next.tasks=next.tasks.map(t=>t.id===agent.taskId?{...t,status:"done" as const}:t);
    next=addEvent(next,agent.name+" completed their task",agent.id);
   }
  }
 }
 if(next.tick % 8 === 1){
  const task=next.tasks.find(t=>t.status==="queued");
  const available=next.agents.find(a=>a.status==="idle"||a.status==="completed");
  if(task&&available)next=assignTask(next,task.id,available.id);
 }
 return next;
}
export function setSimulationSpeed(state:SimState,speed:number):SimState{return {...state,speed:Math.max(0.25,Math.min(4,speed))};}
