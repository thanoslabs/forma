export type RuntimeRef={connectionId:string;machineId?:string;terminalId:string};
export type ExecutionState="pending"|"claimed"|"running"|"blocked"|"needs_review"|"completed"|"failed"|"released";
export type Execution={connectionId:string;schemaVersion:number;id:string;attempt:number;state:ExecutionState;reportedState?:ExecutionState;work:{connectionId:string;projectId:string;itemId:string};prompt?:string;claim?:{id:string;expiresAt?:string};parent?:RuntimeRef;helpers:RuntimeRef[];results:{kind:string;url:string}[];createdAt:string;updatedAt:string};

export async function listExecutions(workspaceId:string,signal?:AbortSignal):Promise<Execution[]>{
 const response=await fetch(`/api/v1/workspaces/${encodeURIComponent(workspaceId)}/executions`,{signal});
 if(!response.ok)throw new Error((await response.text()).trim()||"Unable to load bridge executions");
 return response.json() as Promise<Execution[]>;
}
