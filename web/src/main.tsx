import React from "react";
import { createRoot } from "react-dom/client";
import "./style.css";
import "./office.css";
import "./demo.css";
import { Office } from "./Office";
import { OfficeCharacters } from "./Characters3D";
import { usePlayerMovement } from "./playerMovement";
import { DemoData } from "./DemoData";
import { useSimulation } from "./useSimulation";
import { OfficePanels } from "./OfficePanels";
import { ChatPanel, type ChatMessage } from "./ChatPanel";
import { SpeechBubbles } from "./SpeechBubbles";
import type { SpeechBubble } from "./speech";
import type { LinkedMention } from "./mentions";
import "./panels.css";
import "./layout.css";
import { LayoutEditor } from "./LayoutEditor";
import { defaultLayout, loadLayout, resetLayout, saveLayout } from "./layout";

const sections = [
  { name: "Office", description: "An interactive simulation of your agent workspace. All activity is fictional." },
  { name: "Agents", description: "Connect a harness to discover agents and inspect their sessions." },
  { name: "Projects", description: "Connect Overwatch or another project provider to view work." },
];

function App() {
  const [selected, setSelected] = React.useState(0);
  const { state, dispatch } = useSimulation();
  const [selectedAgent,setSelectedAgent] = React.useState<string | undefined>();
  const [layout,setLayout] = React.useState(loadLayout);
  const playerState=usePlayerMovement(layout.desks,selected===0);
  React.useEffect(()=>saveLayout(layout),[layout]);
  const [messages,setMessages]=React.useState<ChatMessage[]>([{id:0,author:"Forma",body:"Welcome to the office! This is a local demo; messages are not sent to real agents.",mentions:[],time:""}]);
  const [bubbles,setBubbles]=React.useState<SpeechBubble[]>([]);
  const [focusedMessageId,setFocusedMessageId]=React.useState<string>();
  const nextMessageId=React.useRef(1),seenEvent=React.useRef(0);
  const playerPosition=React.useRef(playerState.position);
  playerPosition.current=playerState.position;
  const timers=React.useRef(new Map<string,ReturnType<typeof setTimeout>>());
  const showBubble=React.useCallback((bubble:SpeechBubble)=>{
    setBubbles(old=>[...old.filter(b=>b.speakerId!==bubble.speakerId),bubble]);
    const previous=timers.current.get(bubble.speakerId);if(previous)clearTimeout(previous);
    timers.current.set(bubble.speakerId,setTimeout(()=>{setBubbles(old=>old.filter(b=>b.id!==bubble.id));timers.current.delete(bubble.speakerId);},9000));
  },[]);
  React.useEffect(()=>()=>{timers.current.forEach(clearTimeout);timers.current.clear();},[]);
  React.useEffect(()=>{
    const newEvents=state.events.filter(e=>e.id>seenEvent.current).sort((a,b)=>a.id-b.id);
    for(const event of newEvents){
      seenEvent.current=event.id;
      if(!event.agentId)continue;
      const agent=state.agents.find(a=>a.id===event.agentId);if(!agent)continue;
      const id=nextMessageId.current++;
      const body=event.message;
      setMessages(old=>[...old,{id,author:agent.name,speakerId:agent.id,body,mentions:[],time:""}]);
      showBubble({id:"event-"+event.id,messageId:String(id),speakerId:agent.id,speakerName:agent.name,body,position:agent.position});
    }
  },[state.events,state.agents,showBubble]);
  const sendChat=(body:string,mentions:LinkedMention[])=>{
    const id=nextMessageId.current++;
    setMessages(old=>[...old,{id,author:"You",speakerId:"player",body,mentions,time:new Date().toLocaleTimeString([],{hour:"2-digit",minute:"2-digit"})}]);
    showBubble({id:"chat-"+id,messageId:String(id),speakerId:"player",speakerName:"You",body,position:{...playerPosition.current}});
  };
  const openChatMessage=(messageId:string)=>{
    setFocusedMessageId(messageId);
    requestAnimationFrame(()=>{
      document.getElementById("chat-message-"+messageId)?.scrollIntoView({behavior:"smooth",block:"center"});
    });
  };
  return <div className="app">
    <aside className="sidebar"><h1>forma<span>.</span></h1><p className="eyebrow">WORKSPACE</p>
      <nav aria-label="Main navigation">{sections.map((section, i) =>
        <button key={section.name} className={selected === i ? "active" : ""} onClick={() => setSelected(i)}>{section.name}</button>)}</nav>
      <p className="footnote">Interactive demo · No real agents or tasks</p>
    </aside>
    <main><header><span>Workspace / {sections[selected].name}</span><span className="status">● Demo mode · simulated activity</span></header>
      <section className="content"><p className="eyebrow">WELCOME TO FORMA</p><h2>{sections[selected].name}</h2>
      <p>{sections[selected].description}</p>{selected === 0 ? <><div className="sim-controls"><button onClick={() => dispatch({type:"toggle"})}>{state.running ? "Pause" : "Resume"}</button><button onClick={() => dispatch({type:"reset"})}>Reset</button><label>Speed <select value={state.speed} onChange={e => dispatch({type:"speed",value:Number(e.target.value)})}><option value={0.5}>0.5×</option><option value={1}>1×</option><option value={2}>2×</option><option value={4}>4×</option></select></label></div><p className="player-hint">Explore the office with <kbd>W</kbd><kbd>A</kbd><kbd>S</kbd><kbd>D</kbd> or arrow keys. Desks block movement.</p><div className="office-3d-stack"><Office hideActorArt agents={state.agents} deskPositions={layout.desks} selectedId={selectedAgent} onSelect={setSelectedAgent} onMove={(agentId,p) => dispatch({type:"move",agentId,x:p.x,y:p.y})}/><OfficeCharacters agents={state.agents} playerState={playerState}/><SpeechBubbles bubbles={bubbles} onOpen={openChatMessage}/></div><LayoutEditor layout={layout} onChange={setLayout} onReset={()=>{resetLayout();setLayout(defaultLayout);}}/><ChatPanel agents={state.agents} messages={messages} onSend={sendChat} focusedMessageId={focusedMessageId}/><OfficePanels state={state} selectedId={selectedAgent} onSelect={setSelectedAgent} onAssign={(taskId,agentId)=>dispatch({type:"assign",taskId,agentId})} /></> : <DemoData kind={selected === 1 ? "agents" : "projects"} /> }</section>
    </main>
  </div>;
}

createRoot(document.getElementById("root")!).render(<React.StrictMode><App/></React.StrictMode>);
