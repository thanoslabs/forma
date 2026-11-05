export type AgentProjection={connectionId:string;adapterId:string;id:string;name:string;status:"unknown"|"idle"|"running"|"blocked"|"completed"|"failed";lastSeen:string};

export async function listWorkspaceAgents(workspaceId:string,signal?:AbortSignal):Promise<AgentProjection[]>{
 const response=await fetch(`/api/v1/workspaces/${encodeURIComponent(workspaceId)}/agents`,{signal});
 if(!response.ok)throw new Error((await response.text()).trim()||"Unable to load agents");
 return response.json() as Promise<AgentProjection[]>;
}
