export type ProjectProjection={connectionId:string;id:string;name:string;url?:string};
export type WorkItem={id:string;projectId:string;title:string;status:string;url?:string;updatedAt:string};

async function read<T>(path:string,signal?:AbortSignal):Promise<T>{const response=await fetch(path,{signal});if(!response.ok)throw new Error((await response.text()).trim()||"Unable to load projects");return response.json() as Promise<T>}
export const listProjects=(workspaceId:string,signal?:AbortSignal)=>read<ProjectProjection[]>(`/api/v1/workspaces/${encodeURIComponent(workspaceId)}/projects`,signal);
export const listWorkItems=(workspaceId:string,project:ProjectProjection,signal?:AbortSignal)=>read<WorkItem[]>(`/api/v1/workspaces/${encodeURIComponent(workspaceId)}/connections/${encodeURIComponent(project.connectionId)}/projects/${encodeURIComponent(project.id)}/items`,signal);
