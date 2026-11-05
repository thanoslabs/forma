import { useEffect, useReducer } from "react";
import { addEvent, assignTask, initialSimulation, setSimulationSpeed, tickSimulation, type SimAgent, type SimState } from "./simulation";
import type {AgentProjection} from "./agentApi";
import{createWorkstations}from"./workstations";
import{defaultLayout}from"./layout";
type Action={type:"tick"}|{type:"toggle"}|{type:"speed";value:number}|{type:"assign";taskId:string;agentId:string}|{type:"reset"}|{type:"move";agentId:string;x:number;y:number}|{type:"projectAgents";agents:AgentProjection[]};
const colors=[0x7e70cf,0xe58d9a,0x6abbb1,0xe3ad62,0x73a7d8];
const projectedStatus=(status:AgentProjection["status"]):SimAgent["status"]=>status==="running"?"walking":status==="completed"?"completed":status==="idle"?"idle":"blocked";
function reducer(state:SimState,action:Action):SimState{
 switch(action.type){
  case "tick":return tickSimulation(state);
  case "toggle":return {...state,running:!state.running};
  case "speed":return setSimulationSpeed(state,action.value);
  case "assign":return assignTask(state,action.taskId,action.agentId);
  case "reset":return initialSimulation();
  case "move":return addEvent({...state,agents:state.agents.map(a=>a.id===action.agentId?{...a,position:{x:action.x,y:action.y},target:{x:action.x,y:action.y}}:a)},"Moved "+action.agentId,action.agentId);
  case "projectAgents":return {...state,source:"live",running:true,agents:action.agents.map((agent,index)=>{const id=agent.connectionId+":"+agent.id,current=state.agents.find(item=>item.id===id),station=createWorkstations(defaultLayout.desks)[index%defaultLayout.desks.length],active=agent.status==="running";if(active&&current?.pose==="work")return{...current,name:agent.name};const position=current?.position??{x:192.5+(index%3)*245,y:367.5+Math.floor(index/3)*70};return{id,name:agent.name,status:projectedStatus(agent.status),position,target:active?station.approach:{...position},progress:agent.status==="completed"?100:0,color:current?.color??colors[index%colors.length],providerId:agent.id,connectionId:agent.connectionId,adapterId:agent.adapterId,workstationId:active?station.id:undefined,pose:"stand"};})};
 }
}
export function useSimulation(){
 const [state,dispatch]=useReducer(reducer,undefined,initialSimulation);
 useEffect(()=>{if(!state.running)return;const timer=window.setInterval(()=>dispatch({type:"tick"}),500/state.speed);return()=>window.clearInterval(timer);},[state.running,state.speed]);
 return {state,dispatch};
}
