import { useEffect, useReducer } from "react";
import { addEvent, assignTask, initialSimulation, setSimulationSpeed, tickSimulation, type SimState } from "./simulation";
import { loadSimulation, saveSimulation, clearSimulation } from "./simulationStorage";
type Action={type:"tick"}|{type:"toggle"}|{type:"speed";value:number}|{type:"assign";taskId:string;agentId:string}|{type:"reset"}|{type:"move";agentId:string;x:number;y:number};
function reducer(state:SimState,action:Action):SimState{
 switch(action.type){
  case "tick":return tickSimulation(state);
  case "toggle":return {...state,running:!state.running};
  case "speed":return setSimulationSpeed(state,action.value);
  case "assign":return assignTask(state,action.taskId,action.agentId);
  case "reset":clearSimulation();return initialSimulation();
  case "move":return addEvent({...state,agents:state.agents.map(a=>a.id===action.agentId?{...a,position:{x:action.x,y:action.y},target:{x:action.x,y:action.y}}:a)},"Moved "+action.agentId,action.agentId);
 }
}
export function useSimulation(){
 const [state,dispatch]=useReducer(reducer,undefined,loadSimulation);
 useEffect(()=>{saveSimulation(state);},[state]);
 useEffect(()=>{if(!state.running)return;const timer=window.setInterval(()=>dispatch({type:"tick"}),500/state.speed);return()=>window.clearInterval(timer);},[state.running,state.speed]);
 return {state,dispatch};
}
