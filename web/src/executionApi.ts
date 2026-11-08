export type RuntimeRef={connectionId:string;machineId?:string;terminalId:string};
export type ExecutionState="pending"|"claimed"|"running"|"blocked"|"needs_review"|"completed"|"failed"|"released";
export type Execution={connectionId:string;schemaVersion:number;id:string;attempt:number;state:ExecutionState;reportedState?:ExecutionState;work:{connectionId:string;projectId:string;itemId:string};prompt?:string;claim?:{id:string;expiresAt?:string};parent?:RuntimeRef;helpers:RuntimeRef[];results:{kind:string;url:string}[];createdAt:string;updatedAt:string};
export type ExecutionEvent={schemaVersion:number;id:string;executionId:string;type:"execution.changed"|"runtime.changed"|"helper.linked"|"result.linked";sequence:number;observedAt:string;payload:Record<string,unknown>};

export async function listExecutions(workspaceId:string,signal?:AbortSignal):Promise<Execution[]>{
 const response=await fetch(`/api/v1/workspaces/${encodeURIComponent(workspaceId)}/executions`,{signal});
 if(!response.ok)throw new Error((await response.text()).trim()||"Unable to load bridge executions");
 return response.json() as Promise<Execution[]>;
}
export async function listExecutionEvents(workspaceId:string,connectionId:string,executionId:string,after=0,signal?:AbortSignal):Promise<{events:ExecutionEvent[];next:number}>{
 const path=`/api/v1/workspaces/${encodeURIComponent(workspaceId)}/connections/${encodeURIComponent(connectionId)}/executions/${encodeURIComponent(executionId)}/events?after=${after}`;
 const response=await fetch(path,{signal});if(!response.ok)throw new Error((await response.text()).trim()||"Unable to load execution events");return response.json() as Promise<{events:ExecutionEvent[];next:number}>;
}
