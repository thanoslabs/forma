import React from "react";
import { createRoot } from "react-dom/client";
import "./style.css";
import "./office.css";
import "./demo.css";
import { Office } from "./Office";
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
import "./elevator.css";
import { LayoutEditor } from "./LayoutEditor";
import { defaultLayout, loadLayout, resetLayout, saveLayout } from "./layout";
import { ensureDefaultConversation,ensureWorkspaceConversation,listWorkspaces, listConversationMessages, sendConversationMessage, watchConversationMessages, type ConversationMessage, type ConversationScope,type Workspace } from "./conversationApi";
import {listWorkspaceAgents} from "./agentApi";
import{ELEVATOR_TILE,sceneSpawn,workspaceScenes}from"./scenes";
import{tileKey,worldToTile}from"./worldGrid";

type WorkspacePanel = "chat" | "agents" | "projects" | "activity" | "settings"|"elevator" | null;

function App() {
  const [panel,setPanel] = React.useState<WorkspacePanel>(null);
  const [layout,setLayout] = React.useState(loadLayout);
  const { state, dispatch } = useSimulation();
  const [selectedAgent,setSelectedAgent] = React.useState<string | undefined>();
  const [workspaces,setWorkspaces]=React.useState<Workspace[]>([]),[activeWorkspaceId,setActiveWorkspaceId]=React.useState<string>();
  const scenes=React.useMemo(()=>workspaceScenes(workspaces),[workspaces]),activeScene=scenes.find(scene=>scene.workspaceId===activeWorkspaceId)??scenes[0];
  const occupiedSeats=React.useMemo(()=>new Set(state.agents.filter(agent=>agent.pose==="work"||agent.pose==="sit").map(agent=>agent.workstationId).filter((id):id is string=>!!id)),[state.agents]);
  const playerState=usePlayerMovement(layout.desks,panel===null,occupiedSeats,activeScene?.id??"loading",activeScene?sceneSpawn(activeScene):undefined);
  React.useEffect(()=>saveLayout(layout),[layout]);
  const [messages,setMessages]=React.useState<ChatMessage[]>([{id:"welcome",author:"Forma",body:"Welcome to the shared office conversation. Messages are persisted when the Forma service is connected.",mentions:[],time:""}]);
  const [conversationScope,setConversationScope]=React.useState<ConversationScope>();
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
    const controller=new AbortController();void ensureDefaultConversation(controller.signal).then(async scope=>{const found=await listWorkspaces(controller.signal);setWorkspaces(found);setActiveWorkspaceId(scope.workspaceId);}).catch(()=>{if(!controller.signal.aborted){setChatStatus("local");setChatError("The Forma API is unavailable. Messages will stay in this browser session.");}});return()=>controller.abort();
  },[]);
  React.useEffect(()=>{
    if(!activeWorkspaceId)return;
    const controller=new AbortController();let active=true;let stopWatching:undefined|(()=>void);
    const upsert=(message:ConversationMessage)=>setMessages(old=>[...old.filter(existing=>existing.id!==message.id),toChatMessage(message)]);
    void ensureWorkspaceConversation(activeWorkspaceId,controller.signal).then(async scope=>{
      const found=await listConversationMessages(scope,controller.signal);if(!active)return;
      setConversationScope(scope);setMessages(found.length?found.map(toChatMessage):[]);setChatStatus("connected");setChatError(undefined);
      stopWatching=watchConversationMessages(scope,upsert,()=>{if(active)setChatError("Live updates disconnected. Persisted messages can still be sent.");});
    }).catch(()=>{if(active&&!controller.signal.aborted){setChatStatus("local");setChatError("The Forma API is unavailable. Messages will stay in this browser session.");}});
    return()=>{active=false;controller.abort();stopWatching?.();};
  },[activeWorkspaceId,toChatMessage]);
  React.useEffect(()=>{
    if(!conversationScope)return;
    const controller=new AbortController();
    const refresh=()=>void listWorkspaceAgents(conversationScope.workspaceId,controller.signal).then(agents=>dispatch({type:"projectAgents",agents})).catch(()=>{});
    refresh();const timer=window.setInterval(refresh,5000);return()=>{controller.abort();window.clearInterval(timer);};
  },[conversationScope,dispatch]);
  React.useEffect(()=>{if(tileKey(worldToTile(playerState.position))===tileKey(ELEVATOR_TILE)&&!playerState.seated)setPanel("elevator");},[playerState.position,playerState.seated]);
  const selectFloor=(workspaceId:string)=>{setActiveWorkspaceId(workspaceId);setConversationScope(undefined);setSelectedAgent(undefined);setMessages([]);setBubbles([]);dispatch({type:"projectAgents",agents:[]});setChatStatus("connecting");setPanel(null);};
  const sendChat=(body:string,mentions:LinkedMention[])=>{
    const requestId=globalThis.crypto?.randomUUID?.()??"request-"+Date.now();
    if(chatStatus==="local"||!conversationScope){
      const message:ChatMessage={id:"local-"+requestId,author:"You",speakerId:"player",body,mentions,time:new Date().toLocaleTimeString([],{hour:"2-digit",minute:"2-digit"})};
      setMessages(old=>[...old,message]);showBubble({id:"chat-"+message.id,messageId:message.id,speakerId:"player",speakerName:"You",body,mentions,position:{...playerPosition.current}});return;
    }
    setChatStatus("sending");setChatError(undefined);
    void sendConversationMessage(conversationScope,body,mentions).then(created=>{
      const mapped=toChatMessage(created);setMessages(old=>[...old.filter(existing=>existing.id!==created.id),mapped]);
      showBubble({id:"chat-"+mapped.id,messageId:mapped.id,speakerId:"player",speakerName:mapped.author,body:mapped.body,mentions:mapped.mentions,position:{...playerPosition.current}});
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
        <Office agents={state.agents} playerState={playerState} deskPositions={layout.desks} selectedId={selectedAgent} onSelect={setSelectedAgent} onOpenProjects={()=>setPanel("projects")} onOpenElevator={()=>setPanel("elevator")} onDeselect={()=>setSelectedAgent(undefined)} onMove={(agentId,p)=>dispatch({type:"move",agentId,x:p.x,y:p.y})}/>
        <SpeechBubbles bubbles={bubbles} playerPosition={playerState.position} agents={state.agents} onOpen={openChatMessage}/>
      </div>
    </div>
    <div className="office-hud">
      <div className="office-brand"><strong>forma<span>.</span></strong><small>YOUR AGENT WORKSPACE</small></div>
      <div className="office-status"><span className="office-status-dot"/> {activeScene?`Floor ${activeScene.floor} · ${activeScene.name}`:state.source==="live"?"Workspace · live agent projections":"Demo workspace · simulated activity"}</div>
      <nav className="office-dock" aria-label="Workspace tools">
        {(["chat","agents","projects","activity","settings"] as const).map(item=><button type="button" key={item} aria-pressed={panel===item} className={panel===item?"active":""} onClick={()=>togglePanel(item)}>{item==="chat"?"◉":item==="agents"?"♟":item==="projects"?"▦":item==="activity"?"◷":"⚙"}<span>{item[0].toUpperCase()+item.slice(1)}</span></button>)}
      </nav>
      <div className="office-help">WASD / arrows to move <span>·</span> Click to select · Drag agents to move · Click project board <span>·</span> Esc to close</div>
    </div>
    {panel&&<div className="office-drawer" role="region" aria-label={panel+" panel"}>
      <div className="office-drawer-head"><div><small>{panel==="elevator"?"BUILDING":"WORKSPACE"}</small><h2>{panel[0].toUpperCase()+panel.slice(1)}</h2></div><button type="button" onClick={()=>setPanel(null)} aria-label="Close panel">×</button></div>
      <div className="office-drawer-content">
        {panel==="chat"&&<ChatPanel agents={state.agents} messages={messages} onSend={sendChat} focusedMessageId={focusedMessageId} status={chatStatus} error={chatError}/>}
        {panel==="agents"&&<OfficePanels state={{...state,tasks:[],events:[]}} selectedId={selectedAgent} onSelect={setSelectedAgent} onAssign={(taskId,agentId)=>dispatch({type:"assign",taskId,agentId})}/>}
        {panel==="projects"&&<div className="office-panel-section"><h3>Projects & tasks</h3><p>Demo tasks in your office. Connect a project provider for live projects.</p>{state.tasks.map(task=><article className="office-project" key={task.id}><strong>{task.title}</strong><span>{task.status}</span>{task.status==="queued"&&<select aria-label={"Assign "+task.title} defaultValue="" onChange={e=>{if(e.target.value)dispatch({type:"assign",taskId:task.id,agentId:e.target.value});}}><option value="">Assign agent…</option>{state.agents.filter(a=>a.status==="idle"||a.status==="completed").map(a=><option key={a.id} value={a.id}>{a.name}</option>)}</select>}</article>)}</div>}
        {panel==="activity"&&<div className="office-panel-section"><h3>Office activity</h3><p>Simulation events are separate from chat.</p><ol className="office-activity">{state.events.map(event=><li key={event.id}><small>Tick {event.tick}</small><span>{event.message}</span></li>)}</ol></div>}
        {panel==="settings"&&<div className="office-panel-section"><h3>Simulation</h3><div className="office-setting"><button onClick={()=>dispatch({type:"toggle"})}>{state.running?"Pause simulation":"Resume simulation"}</button><button onClick={()=>dispatch({type:"reset"})}>Reset simulation</button></div><label className="office-speed">Simulation speed <select value={state.speed} onChange={e=>dispatch({type:"speed",value:Number(e.target.value)})}><option value={0.5}>0.5×</option><option value={1}>1×</option><option value={2}>2×</option><option value={4}>4×</option></select></label><LayoutEditor layout={layout} onChange={setLayout} onReset={()=>{resetLayout();setLayout(defaultLayout);}}/></div>}
        {panel==="elevator"&&<div className="office-panel-section"><h3>Select a floor</h3><p>Each floor is a Forma workspace with its own agents and conversation.</p><div className="elevator-floors">{scenes.map(scene=><button key={scene.id} className={scene.workspaceId===activeWorkspaceId?"active":""} onClick={()=>selectFloor(scene.workspaceId!)}><strong>{scene.floor}</strong><span>{scene.name}</span></button>)}</div></div>}
      </div>
    </div>}
  </div>;
}

createRoot(document.getElementById("root")!).render(<React.StrictMode><App/></React.StrictMode>);
