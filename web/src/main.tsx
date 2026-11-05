import React from "react";
import { createRoot } from "react-dom/client";
import "./style.css";
import "./office.css";
import "./demo.css";
import { Office } from "./Office";
import { OfficeCharacters } from "./Characters3D";
import { usePlayerMovement } from "./playerMovement";
import { useSimulation } from "./useSimulation";
import { OfficePanels } from "./OfficePanels";
import { ChatPanel, type ChatMessage } from "./ChatPanel";
import { SpeechBubbles } from "./SpeechBubbles";
import type { SpeechBubble } from "./speech";
import type { LinkedMention } from "./mentions";
import "./panels.css";
import "./layout.css";
import "./immersive.css";
import { LayoutEditor } from "./LayoutEditor";
import { defaultLayout, loadLayout, resetLayout, saveLayout } from "./layout";
import { listConversationMessages, sendConversationMessage, type ConversationMessage } from "./conversationApi";

type WorkspacePanel = "chat" | "agents" | "projects" | "activity" | "settings" | null;

function App() {
  const [panel,setPanel] = React.useState<WorkspacePanel>(null);
  const { state, dispatch } = useSimulation();
  const [selectedAgent,setSelectedAgent] = React.useState<string | undefined>();
  const [layout,setLayout] = React.useState(loadLayout);
  const playerState=usePlayerMovement(layout.desks,panel===null);
  React.useEffect(()=>saveLayout(layout),[layout]);
  const [messages,setMessages]=React.useState<ChatMessage[]>([{id:"welcome",author:"Forma",body:"Welcome to the office. Mention Atlas, Nova, or Pixel to send a message through the mock agent harness.",mentions:[],time:""}]);
  const [chatStatus,setChatStatus]=React.useState<"connecting"|"connected"|"sending"|"local">("connecting");
  const [chatError,setChatError]=React.useState<string>();
  const [bubbles,setBubbles]=React.useState<SpeechBubble[]>([]);
  const [focusedMessageId,setFocusedMessageId]=React.useState<string>();
  const playerPosition=React.useRef(playerState.position);
  playerPosition.current=playerState.position;
  const timers=React.useRef(new Map<string,ReturnType<typeof setTimeout>>());
  const showBubble=React.useCallback((bubble:SpeechBubble)=>{
    setBubbles(old=>[...old.filter(b=>b.speakerId!==bubble.speakerId),bubble]);
    const previous=timers.current.get(bubble.speakerId);if(previous)clearTimeout(previous);
    timers.current.set(bubble.speakerId,setTimeout(()=>{setBubbles(old=>old.filter(b=>b.id!==bubble.id));timers.current.delete(bubble.speakerId);},9000));
  },[]);
  React.useEffect(()=>()=>{timers.current.forEach(clearTimeout);timers.current.clear();},[]);
  const toChatMessage=React.useCallback((message:ConversationMessage):ChatMessage=>({id:message.id,author:message.author.name,speakerId:message.author.kind==="agent"?message.author.id:"player",body:message.body,mentions:message.mentions,time:new Date(message.occurredAt).toLocaleTimeString([],{hour:"2-digit",minute:"2-digit"})}),[]);
  React.useEffect(()=>{
    const controller=new AbortController();
    void listConversationMessages(controller.signal).then(found=>{
      if(found.length)setMessages(found.map(toChatMessage));
      setChatStatus("connected");setChatError(undefined);
    }).catch(()=>{if(!controller.signal.aborted){setChatStatus("local");setChatError("The demo API is unavailable. Messages will stay in this browser session.");}});
    return()=>controller.abort();
  },[toChatMessage]);
  const sendChat=(body:string,mentions:LinkedMention[])=>{
    const requestId=globalThis.crypto?.randomUUID?.()??"request-"+Date.now();
    if(chatStatus==="local"){
      const message:ChatMessage={id:"local-"+requestId,author:"You",speakerId:"player",body,mentions,time:new Date().toLocaleTimeString([],{hour:"2-digit",minute:"2-digit"})};
      setMessages(old=>[...old,message]);showBubble({id:"chat-"+message.id,messageId:message.id,speakerId:"player",speakerName:"You",body,mentions,position:{...playerPosition.current}});return;
    }
    setChatStatus("sending");setChatError(undefined);
    void sendConversationMessage(body,mentions,requestId).then(created=>{
      const mapped=created.map(toChatMessage);
      setMessages(old=>[...old.filter(existing=>!created.some(message=>message.id===existing.id)),...mapped]);
      for(const message of mapped){
        const agent=state.agents.find(item=>item.id===message.speakerId);
        showBubble({id:"chat-"+message.id,messageId:message.id,speakerId:message.speakerId??"player",speakerName:message.author,body:message.body,mentions:message.mentions,position:agent?{...agent.position}:{...playerPosition.current}});
      }
      setChatStatus("connected");
    }).catch(error=>{
      const fallback:ChatMessage={id:"local-"+requestId,author:"You",speakerId:"player",body,mentions,time:new Date().toLocaleTimeString([],{hour:"2-digit",minute:"2-digit"})};
      setMessages(old=>[...old,fallback]);showBubble({id:"chat-"+fallback.id,messageId:fallback.id,speakerId:"player",speakerName:"You",body,mentions,position:{...playerPosition.current}});
      setChatStatus("local");setChatError(error instanceof Error?error.message:"The demo API is unavailable. Message kept locally.");
    });
  };
  const openChatMessage=(messageId:string)=>{
    setFocusedMessageId(messageId);
    requestAnimationFrame(()=>{
      document.getElementById("chat-message-"+messageId)?.scrollIntoView({behavior:"smooth",block:"center"});
    });
  };
  const togglePanel=(next:Exclude<WorkspacePanel,null>)=>setPanel(current=>current===next?null:next);
  React.useEffect(()=>{
    const onKey=(event:KeyboardEvent)=>{if(event.key==="Escape")setPanel(null);};
    window.addEventListener("keydown",onKey);return()=>window.removeEventListener("keydown",onKey);
  },[]);
  return <div className="office-app">
    <div className="office-world" aria-label="Forma office workspace">
      <div className="office-world-stage">
        <Office hideActorArt agents={state.agents} deskPositions={layout.desks} selectedId={selectedAgent} onSelect={setSelectedAgent} onOpenProjects={()=>setPanel("projects")} onDeselect={()=>setSelectedAgent(undefined)} onMove={(agentId,p)=>dispatch({type:"move",agentId,x:p.x,y:p.y})}/>
        <OfficeCharacters agents={state.agents} playerState={playerState} selectedId={selectedAgent}/>
        <SpeechBubbles bubbles={bubbles} playerPosition={playerState.position} agents={state.agents} onOpen={openChatMessage}/>
      </div>
    </div>
    <div className="office-hud">
      <div className="office-brand"><strong>forma<span>.</span></strong><small>YOUR AGENT WORKSPACE</small></div>
      <div className="office-status"><span className="office-status-dot"/> Demo workspace · simulated activity</div>
      <nav className="office-dock" aria-label="Workspace tools">
        {(["chat","agents","projects","activity","settings"] as const).map(item=><button type="button" key={item} aria-pressed={panel===item} className={panel===item?"active":""} onClick={()=>togglePanel(item)}>{item==="chat"?"◉":item==="agents"?"♟":item==="projects"?"▦":item==="activity"?"◷":"⚙"}<span>{item[0].toUpperCase()+item.slice(1)}</span></button>)}
      </nav>
      <div className="office-help">WASD / arrows to move <span>·</span> Click to select · Drag agents to move · Click project board <span>·</span> Esc to close</div>
    </div>
    {panel&&<div className="office-drawer" role="region" aria-label={panel+" panel"}>
      <div className="office-drawer-head"><div><small>WORKSPACE</small><h2>{panel[0].toUpperCase()+panel.slice(1)}</h2></div><button type="button" onClick={()=>setPanel(null)} aria-label="Close panel">×</button></div>
      <div className="office-drawer-content">
        {panel==="chat"&&<ChatPanel agents={state.agents} messages={messages} onSend={sendChat} focusedMessageId={focusedMessageId} status={chatStatus} error={chatError}/>}
        {panel==="agents"&&<OfficePanels state={{...state,tasks:[],events:[]}} selectedId={selectedAgent} onSelect={setSelectedAgent} onAssign={(taskId,agentId)=>dispatch({type:"assign",taskId,agentId})}/>}
        {panel==="projects"&&<div className="office-panel-section"><h3>Projects & tasks</h3><p>Demo tasks in your office. Connect a project provider for live projects.</p>{state.tasks.map(task=><article className="office-project" key={task.id}><strong>{task.title}</strong><span>{task.status}</span>{task.status==="queued"&&<select aria-label={"Assign "+task.title} defaultValue="" onChange={e=>{if(e.target.value)dispatch({type:"assign",taskId:task.id,agentId:e.target.value});}}><option value="">Assign agent…</option>{state.agents.filter(a=>a.status==="idle"||a.status==="completed").map(a=><option key={a.id} value={a.id}>{a.name}</option>)}</select>}</article>)}</div>}
        {panel==="activity"&&<div className="office-panel-section"><h3>Office activity</h3><p>Simulation events are separate from chat.</p><ol className="office-activity">{state.events.map(event=><li key={event.id}><small>Tick {event.tick}</small><span>{event.message}</span></li>)}</ol></div>}
        {panel==="settings"&&<div className="office-panel-section"><h3>Simulation</h3><div className="office-setting"><button onClick={()=>dispatch({type:"toggle"})}>{state.running?"Pause simulation":"Resume simulation"}</button><button onClick={()=>dispatch({type:"reset"})}>Reset simulation</button></div><label className="office-speed">Simulation speed <select value={state.speed} onChange={e=>dispatch({type:"speed",value:Number(e.target.value)})}><option value={0.5}>0.5×</option><option value={1}>1×</option><option value={2}>2×</option><option value={4}>4×</option></select></label><LayoutEditor layout={layout} onChange={setLayout} onReset={()=>{resetLayout();setLayout(defaultLayout);}}/></div>}
      </div>
    </div>}
  </div>;
}

createRoot(document.getElementById("root")!).render(<React.StrictMode><App/></React.StrictMode>);
