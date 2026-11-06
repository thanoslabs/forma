import React from"react";
import{getAgentDetail,type AgentDetail}from"./agentApi";
import type{SimAgent}from"./simulation";

type Props={workspaceId:string;agent:SimAgent;onMessage:(body:string)=>void};
export function AgentInspector({workspaceId,agent,onMessage}:Props){
 const[detail,setDetail]=React.useState<AgentDetail>(),[error,setError]=React.useState<string>(),[body,setBody]=React.useState("");
 React.useEffect(()=>{if(!agent.connectionId||!agent.providerId)return;const controller=new AbortController();setDetail(undefined);setError(undefined);void getAgentDetail(workspaceId,agent.connectionId,agent.providerId,controller.signal).then(setDetail).catch(value=>{if(!controller.signal.aborted)setError(value instanceof Error?value.message:"Unable to load agent")});return()=>controller.abort();},[workspaceId,agent.connectionId,agent.providerId]);
 const canMessage=detail?.capabilities.includes("input")??false;
 return <div className="agent-detail"><h4>{agent.name}</h4><p><span className="connection-health connected"/> {detail?"Connected":"Loading provider…"}</p><dl><div><dt>Status</dt><dd>{detail?.agent.status??agent.status}</dd></div><div><dt>Provider</dt><dd>{agent.adapterId??"simulation"}</dd></div><div><dt>Last seen</dt><dd>{detail?.agent.lastSeen?new Date(detail.agent.lastSeen).toLocaleString():"—"}</dd></div></dl>{error&&<p className="connection-error">{error}</p>}{detail?.capabilities.includes("sessions")&&<section><h5>Sessions</h5>{detail.sessions.length?<ul>{detail.sessions.map(session=><li key={session.id}><strong>{session.id}</strong> · {session.status}</li>)}</ul>:<p>No active sessions.</p>}</section>}{canMessage&&<form onSubmit={event=>{event.preventDefault();const value=body.trim();if(value){onMessage(value);setBody("")}}}><label>Message agent<textarea value={body} onChange={event=>setBody(event.target.value)} placeholder={`Ask ${agent.name}…`}/></label><button type="submit" disabled={!body.trim()}>Send message</button></form>}</div>;
}
