import React from "react";
import { createRoot } from "react-dom/client";
import "./style.css";
import "./office.css";
import "./demo.css";
import { Office } from "./Office";
import { DemoData } from "./DemoData";
import { useSimulation } from "./useSimulation";
import { OfficePanels } from "./OfficePanels";
import "./panels.css";
import "./layout.css";
import { LayoutEditor } from "./LayoutEditor";
import { AvatarCustomizer, palette } from "./AvatarCustomizer";
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
  const [avatarColors,setAvatarColors] = React.useState<Record<string,number>>(()=>{try{const v=JSON.parse(localStorage.getItem("forma:avatar-colors:v1")??"{}") as Record<string,number>;return Object.fromEntries(Object.entries(v).filter(([id,color])=>["atlas","nova","pixel"].includes(id)&&palette.includes(color)));}catch{return {};}});
  React.useEffect(()=>{try{localStorage.setItem("forma:avatar-colors:v1",JSON.stringify(avatarColors));}catch{ /* storage unavailable */ }},[avatarColors]);
  React.useEffect(()=>saveLayout(layout),[layout]);
  return <div className="app">
    <aside className="sidebar"><h1>forma<span>.</span></h1><p className="eyebrow">WORKSPACE</p>
      <nav aria-label="Main navigation">{sections.map((section, i) =>
        <button key={section.name} className={selected === i ? "active" : ""} onClick={() => setSelected(i)}>{section.name}</button>)}</nav>
      <p className="footnote">Interactive demo · No real agents or tasks</p>
    </aside>
    <main><header><span>Workspace / {sections[selected].name}</span><span className="status">● Demo mode · simulated activity</span></header>
      <section className="content"><p className="eyebrow">WELCOME TO FORMA</p><h2>{sections[selected].name}</h2>
      <p>{sections[selected].description}</p>{selected === 0 ? <><div className="sim-controls"><button onClick={() => dispatch({type:"toggle"})}>{state.running ? "Pause" : "Resume"}</button><button onClick={() => dispatch({type:"reset"})}>Reset</button><label>Speed <select value={state.speed} onChange={e => dispatch({type:"speed",value:Number(e.target.value)})}><option value={0.5}>0.5×</option><option value={1}>1×</option><option value={2}>2×</option><option value={4}>4×</option></select></label></div><Office key={JSON.stringify(avatarColors)} agents={state.agents.map(a=>({...a,color:avatarColors[a.id]??a.color}))} deskPositions={layout.desks} selectedId={selectedAgent} onSelect={setSelectedAgent} onMove={(agentId,p) => dispatch({type:"move",agentId,x:p.x,y:p.y})}/><AvatarCustomizer agents={state.agents} colors={avatarColors} onChange={(id,color)=>setAvatarColors(prev=>({...prev,[id]:color}))}/><LayoutEditor layout={layout} onChange={setLayout} onReset={()=>{resetLayout();setLayout(defaultLayout);}}/><OfficePanels state={state} selectedId={selectedAgent} onSelect={setSelectedAgent} onAssign={(taskId,agentId)=>dispatch({type:"assign",taskId,agentId})} /></> : <DemoData kind={selected === 1 ? "agents" : "projects"} /> }</section>
    </main>
  </div>;
}

createRoot(document.getElementById("root")!).render(<React.StrictMode><App/></React.StrictMode>);
