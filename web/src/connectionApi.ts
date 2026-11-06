export type Connection={id:string;workspaceId:string;adapterId:string;providerType:"agent"|"project"|"chat";displayName:string;endpointUrl?:string;createdAt:string};
export type ConnectionCheck={connectionId:string;status:"connected";agentCount:number;capabilities:string[];checkedAt:string};
export type CreateConnection={adapterId:string;providerType:Connection["providerType"];displayName:string;endpointUrl:string;credential:string};

async function request<T>(path:string,init?:RequestInit):Promise<T>{const response=await fetch(path,init);if(!response.ok)throw new Error((await response.text()).trim()||"Connection request failed");return response.json() as Promise<T>;}
const base=(workspaceId:string)=>`/api/v1/workspaces/${encodeURIComponent(workspaceId)}/connections`;
export function listConnections(workspaceId:string,signal?:AbortSignal):Promise<Connection[]>{return request(base(workspaceId),{signal});}
export function createConnection(workspaceId:string,input:CreateConnection):Promise<Connection>{return request(base(workspaceId),{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(input)});}
export function testConnection(workspaceId:string,connectionId:string):Promise<ConnectionCheck>{return request(`${base(workspaceId)}/${encodeURIComponent(connectionId)}/test`,{method:"POST"});}
