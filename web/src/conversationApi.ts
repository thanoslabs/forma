import type {LinkedMention} from "./mentions";

export type ConversationMessage={id:string;conversationId:string;body:string;author:{kind:"user"|"agent"|"system";id:string;name:string};mentions:LinkedMention[];occurredAt:string;replyToMessageId?:string};
export type Workspace={id:string;name:string};
type Conversation={id:string;workspaceId:string;title:string};
export type ConversationScope={workspaceId:string;conversationId:string};

async function request<T>(path:string,init?:RequestInit):Promise<T>{
 const response=await fetch(path,init);if(!response.ok){const message=(await response.text()).trim();throw new Error(message||"Forma API request failed");}return response.json() as Promise<T>;
}
const json=(body:unknown):RequestInit=>({method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(body)});

export async function ensureDefaultConversation(signal?:AbortSignal):Promise<ConversationScope>{
 let workspaces=await listWorkspaces(signal);
 if(!workspaces.length)workspaces=[await request<Workspace>("/api/v1/workspaces",{...json({name:"Forma"}),signal})];
 return ensureWorkspaceConversation(workspaces[0].id,signal);
}
export function listWorkspaces(signal?:AbortSignal):Promise<Workspace[]>{return request<Workspace[]>("/api/v1/workspaces",{signal});}
export async function ensureWorkspaceConversation(workspaceId:string,signal?:AbortSignal):Promise<ConversationScope>{
 let conversations=await request<Conversation[]>(`/api/v1/workspaces/${encodeURIComponent(workspaceId)}/conversations`,{signal});
 if(!conversations.length)conversations=[await request<Conversation>(`/api/v1/workspaces/${encodeURIComponent(workspaceId)}/conversations`,{...json({title:"Office"}),signal})];
 return {workspaceId,conversationId:conversations[0].id};
}

function messagesPath(scope:ConversationScope):string{return `/api/v1/workspaces/${encodeURIComponent(scope.workspaceId)}/conversations/${encodeURIComponent(scope.conversationId)}/messages`;}
export function listConversationMessages(scope:ConversationScope,signal?:AbortSignal):Promise<ConversationMessage[]>{return request(messagesPath(scope),{signal});}
export function sendConversationMessage(scope:ConversationScope,body:string,mentions:LinkedMention[],idempotencyKey:string=crypto.randomUUID()):Promise<ConversationMessage>{
 const init=json({body,mentions});
 init.headers={...(init.headers as Record<string,string>),"Idempotency-Key":idempotencyKey};
 return request(messagesPath(scope),init);
}

export function watchConversationMessages(scope:ConversationScope,onMessage:(message:ConversationMessage)=>void,onDisconnect:()=>void):()=>void{
 const url=new URL(messagesPath(scope)+"/ws",window.location.href);url.protocol=url.protocol==="https:"?"wss:":"ws:";
 const socket=new WebSocket(url);socket.onmessage=event=>onMessage(JSON.parse(String(event.data)) as ConversationMessage);socket.onerror=onDisconnect;socket.onclose=onDisconnect;
 return()=>socket.close(1000,"client closed");
}
