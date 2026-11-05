import { useEffect, useReducer } from "react";
import { addEvent, assignTask, initialSimulation, setSimulationSpeed, tickSimulation, type SimAgent, type SimState } from "./simulation";
import type {AgentProjection} from "./agentApi";
type Action={type:"tick"}|{type:"toggle"}|{type:"speed";value:number}|{type:"assign";taskId:string;agentId:string}|{type:"reset"}|{type:"move";agentId:string;x:number;y:number}|{type:"projectAgents";agents:AgentProjection[]};
const colors=[0x7e70cf,0xe58d9a,0x6abbb1,0xe3ad62,0x73a7d8];
const projectedStatus=(status:AgentProjection["status"]):SimAgent["status"]=>status==="running"?"working":status==="completed"?"completed":status==="idle"?"idle":"blocked";
function reducer(state:SimState,action:Action):SimState{
 switch(action.type){
  case "tick":return tickSimulation(state);
  case "toggle":return {...state,running:!state.running};
  case "speed":return setSimulationSpeed(state,action.value);
  case "assign":return assignTask(state,action.taskId,action.agentId);
  case "reset":return initialSimulation();
  case "move":return addEvent({...state,agents:state.agents.map(a=>a.id===action.agentId?{...a,position:{x:action.x,y:action.y},target:{x:action.x,y:action.y}}:a)},"Moved "+action.agentId,action.agentId);
  case "projectAgents":return {...state,source:"live",running:false,agents:action.agents.map((agent,index)=>{const id=agent.connectionId+":"+agent.id;const current=state.agents.find(item=>item.id===id);const position=current?.position??{x:180+(index%3)*250,y:350+Math.floor(index/3)*120};return{id,name:agent.name,status:projectedStatus(agent.status),position,target:{...position},progress:agent.status==="completed"?100:0,color:current?.color??colors[index%colors.length],providerId:agent.id,connectionId:agent.connectionId,adapterId:agent.adapterId};})};
 }
}
export function useSimulation(){
 const [state,dispatch]=useReducer(reducer,undefined,initialSimulation);
 useEffect(()=>{if(!state.running)return;const timer=window.setInterval(()=>dispatch({type:"tick"}),500/state.speed);return()=>window.clearInterval(timer);},[state.running,state.speed]);
 return {state,dispatch};
}
