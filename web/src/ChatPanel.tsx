import { useMemo, useRef, useState } from "react";
import { activeMention, insertMention, mentionOptions, resolveMentions, type LinkedMention, type MentionTarget } from "./mentions";
import type { SimAgent } from "./simulation";
import "./chat.css";

export type ChatMessage={id:string;body:string;mentions:LinkedMention[];author:string;time:string;speakerId?:string};
const projects:MentionTarget[]=[
 {kind:"project",id:"forma",name:"Forma"},
 {kind:"project",id:"office-redesign",name:"Office Redesign"},
 {kind:"project",id:"agent-automation",name:"Agent Automation"}
];
export function ChatPanel({agents,messages,onSend,focusedMessageId,status,error}:{agents:SimAgent[];messages:ChatMessage[];onSend:(body:string,mentions:LinkedMention[])=>void;focusedMessageId?:string;status:"connecting"|"connected"|"sending"|"local";error?:string}) {
 const targets=useMemo<MentionTarget[]>(()=>[...agents.map(a=>({kind:"agent" as const,id:a.id,name:a.name})),...projects],[agents]);
 const [draft,setDraft]=useState(""),[caret,setCaret]=useState(0),[choice,setChoice]=useState(0);
 const input=useRef<HTMLTextAreaElement>(null);
 const match=activeMention(draft,caret),options=match?mentionOptions(match,targets):[];
 const select=(target:MentionTarget)=>{
  if(!match)return;
  const next=insertMention(draft,match,target,targets);
  const position=match.start+next.length-(draft.length-(match.end-match.start));
  setDraft(next);setCaret(position);setChoice(0);
  requestAnimationFrame(()=>{input.current?.focus();input.current?.setSelectionRange(position,position);});
 };
 const send=()=>{
  const body=draft.trim();if(!body)return;
  const mentions=resolveMentions(body,targets);
  onSend(body,mentions);
  setDraft("");setCaret(0);setChoice(0);
 };
 const renderMessage=(message:ChatMessage)=>{
  if(!message.mentions.length)return message.body;
  const parts:React.ReactNode[]=[];let offset=0;
  for(const m of message.mentions){
   parts.push(message.body.slice(offset,m.start));
   parts.push(<span key={m.start} className={"chat-linked "+m.kind} title={m.kind+": "+m.name}>{m.text}</span>);
   offset=m.end;
  }
  parts.push(message.body.slice(offset));return parts;
 };
 return <section className="chat-panel" aria-label="Office chat">
  <div className="chat-heading"><div><h3>Office chat</h3><p>Demo conversation · mention agents and projects with @</p></div><span className="chat-demo">{status==="connected"?"MOCK AGENT":status==="sending"?"SENDING":status==="connecting"?"CONNECTING":"LOCAL FALLBACK"}</span></div>
  {error&&<p role="status" className="chat-error">{error}</p>}
  <div className="chat-messages" role="log" aria-live="polite">{messages.map(m=><article key={m.id} id={"chat-message-"+m.id} className={"chat-message "+(m.author==="You"?"mine":"")+(focusedMessageId===String(m.id)?" focused":"")}>
   <div className="chat-message-meta"><strong>{m.author}</strong><small>{m.time}</small></div><p>{renderMessage(m)}</p>
  </article>)}</div>
  <div className="chat-compose">
   {match&&options.length>0&&<div className="chat-suggestions" role="listbox" aria-label="Mention suggestions">{options.map((target,i)=><button type="button" role="option" aria-selected={choice===i} className={choice===i?"active":""} key={target.kind+target.id} onMouseDown={e=>e.preventDefault()} onClick={()=>select(target)}><span className="chat-kind">{target.kind}</span><strong>{target.name}</strong><small>@{target.kind}:{target.id}</small></button>)}</div>}
   <label htmlFor="office-chat-input" className="chat-label">Message</label>
   <textarea id="office-chat-input" ref={input} rows={2} value={draft} placeholder="Message the office… Try @Atlas or @project:Forma" onChange={e=>{setDraft(e.target.value);setCaret(e.target.selectionStart);setChoice(0);}} onClick={e=>setCaret(e.currentTarget.selectionStart)} onKeyUp={e=>{if(!["ArrowDown","ArrowUp","Enter","Escape"].includes(e.key))setCaret(e.currentTarget.selectionStart);}} onKeyDown={e=>{
    if(options.length&&e.key==="ArrowDown"){e.preventDefault();setChoice(i=>(i+1)%options.length);}
    else if(options.length&&e.key==="ArrowUp"){e.preventDefault();setChoice(i=>(i+options.length-1)%options.length);}
    else if(options.length&&(e.key==="Enter"||e.key==="Tab")){e.preventDefault();select(options[choice]??options[0]);}
    else if(e.key==="Enter"&&!e.shiftKey){e.preventDefault();send();}
   }}/>
   <div className="chat-actions"><small>Enter to send · Shift+Enter for newline · @agent:name / @project:name</small><button type="button" disabled={!draft.trim()||status==="sending"} onClick={send}>Send</button></div>
  </div>
 </section>;
}
