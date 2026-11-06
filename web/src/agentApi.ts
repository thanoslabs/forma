export type AgentProjection={connectionId:string;adapterId:string;id:string;name:string;status:"unknown"|"idle"|"running"|"blocked"|"completed"|"failed";lastSeen:string};

export async function listWorkspaceAgents(workspaceId:string,signal?:AbortSignal):Promise<AgentProjection[]>{
 const response=await fetch(`/api/v1/workspaces/${encodeURIComponent(workspaceId)}/agents`,{signal});
 if(!response.ok)throw new Error((await response.text()).trim()||"Unable to load agents");
 return response.json() as Promise<AgentProjection[]>;
}

export type AgentSnapshot={type:"agents.snapshot";workspaceId:string;agents:AgentProjection[];observedAt:string};

export function agentStreamURL(workspaceId:string,location:Pick<Location,"protocol"|"host">=globalThis.location):string{
 const protocol=location.protocol==="https:"?"wss:":"ws:";
 return `${protocol}//${location.host}/api/v1/workspaces/${encodeURIComponent(workspaceId)}/agents/ws`;
}

export function watchWorkspaceAgents(workspaceId:string,onSnapshot:(agents:AgentProjection[])=>void,onDisconnect:()=>void=()=>{}):()=>void{
 const socket=new WebSocket(agentStreamURL(workspaceId));
 let closed=false;
 socket.addEventListener("message",event=>{const snapshot=JSON.parse(String(event.data)) as AgentSnapshot;if(snapshot.type==="agents.snapshot"&&snapshot.workspaceId===workspaceId)onSnapshot(snapshot.agents);});
 socket.addEventListener("close",()=>{if(!closed)onDisconnect();});
 socket.addEventListener("error",()=>socket.close());
 return()=>{closed=true;socket.close();};
}
