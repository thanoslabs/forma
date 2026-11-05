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

const sections = [
  { name: "Office", description: "An interactive simulation of your agent workspace. All activity is fictional." },
  { name: "Agents", description: "Connect a harness to discover agents and inspect their sessions." },
  { name: "Projects", description: "Connect Overwatch or another project provider to view work." },
];

function App() {
  const [panel,setPanel] = React.useState<WorkspacePanel>(null);
  const { state, dispatch } = useSimulation();
  const [selectedAgent,setSelectedAgent] = React.useState<string | undefined>();
  const [layout,setLayout] = React.useState(loadLayout);
  const playerState=usePlayerMovement(layout.desks,panel===null);
  React.useEffect(()=>saveLayout(layout),[layout]);
  const [messages,setMessages]=React.useState<ChatMessage[]>([{id:0,author:"Forma",body:"Welcome to the office! This is a local demo; messages are not sent to real agents.",mentions:[],time:""}]);
  const [bubbles,setBubbles]=React.useState<SpeechBubble[]>([]);
  const [focusedMessageId,setFocusedMessageId]=React.useState<string>();
  const nextMessageId=React.useRef(1);
  const playerPosition=React.useRef(playerState.position);
  playerPosition.current=playerState.position;
  const timers=React.useRef(new Map<string,ReturnType<typeof setTimeout>>());
  const showBubble=React.useCallback((bubble:SpeechBubble)=>{
    setBubbles(old=>[...old.filter(b=>b.speakerId!==bubble.speakerId),bubble]);
    const previous=timers.current.get(bubble.speakerId);if(previous)clearTimeout(previous);
    timers.current.set(bubble.speakerId,setTimeout(()=>{setBubbles(old=>old.filter(b=>b.id!==bubble.id));timers.current.delete(bubble.speakerId);},9000));
  },[]);
  React.useEffect(()=>()=>{timers.current.forEach(clearTimeout);timers.current.clear();},[]);
  const sendChat=(body:string,mentions:LinkedMention[])=>{
    const id=nextMessageId.current++;
    setMessages(old=>[...old,{id,author:"You",speakerId:"player",body,mentions,time:new Date().toLocaleTimeString([],{hour:"2-digit",minute:"2-digit"})}]);
    showBubble({id:"chat-"+id,messageId:String(id),speakerId:"player",speakerName:"You",body,mentions,position:{...playerPosition.current}});
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
        <Office hideActorArt agents={state.agents} deskPositions={layout.desks} selectedId={selectedAgent} onSelect={id=>{setSelectedAgent(id);setPanel("agents");}} onDeselect={()=>setSelectedAgent(undefined)} onMove={(agentId,p)=>dispatch({type:"move",agentId,x:p.x,y:p.y})}/>
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
      <div className="office-help">WASD / arrows to move <span>·</span> Click an agent to inspect <span>·</span> Esc to close</div>
    </div>
    {panel&&<div className="office-drawer" role="region" aria-label={panel+" panel"}>
      <div className="office-drawer-head"><div><small>WORKSPACE</small><h2>{panel[0].toUpperCase()+panel.slice(1)}</h2></div><button type="button" onClick={()=>setPanel(null)} aria-label="Close panel">×</button></div>
      <div className="office-drawer-content">
        {panel==="chat"&&<ChatPanel agents={state.agents} messages={messages} onSend={sendChat} focusedMessageId={focusedMessageId}/>}
        {panel==="agents"&&<OfficePanels state={{...state,tasks:[],events:[]}} selectedId={selectedAgent} onSelect={setSelectedAgent} onAssign={(taskId,agentId)=>dispatch({type:"assign",taskId,agentId})}/>}
        {panel==="projects"&&<div className="office-panel-section"><h3>Projects & tasks</h3><p>Demo tasks in your office. Connect a project provider for live projects.</p>{state.tasks.map(task=><article className="office-project" key={task.id}><strong>{task.title}</strong><span>{task.status}</span>{task.status==="queued"&&<select aria-label={"Assign "+task.title} defaultValue="" onChange={e=>{if(e.target.value)dispatch({type:"assign",taskId:task.id,agentId:e.target.value});}}><option value="">Assign agent…</option>{state.agents.filter(a=>a.status==="idle"||a.status==="completed").map(a=><option key={a.id} value={a.id}>{a.name}</option>)}</select>}</article>)}</div>}
        {panel==="activity"&&<div className="office-panel-section"><h3>Office activity</h3><p>Simulation events are separate from chat.</p><ol className="office-activity">{state.events.map(event=><li key={event.id}><small>Tick {event.tick}</small><span>{event.message}</span></li>)}</ol></div>}
        {panel==="settings"&&<div className="office-panel-section"><h3>Simulation</h3><div className="office-setting"><button onClick={()=>dispatch({type:"toggle"})}>{state.running?"Pause simulation":"Resume simulation"}</button><button onClick={()=>dispatch({type:"reset"})}>Reset simulation</button></div><label className="office-speed">Simulation speed <select value={state.speed} onChange={e=>dispatch({type:"speed",value:Number(e.target.value)})}><option value={0.5}>0.5×</option><option value={1}>1×</option><option value={2}>2×</option><option value={4}>4×</option></select></label><LayoutEditor layout={layout} onChange={setLayout} onReset={()=>{resetLayout();setLayout(defaultLayout);}}/></div>}
      </div>
    </div>}
  </div>;
