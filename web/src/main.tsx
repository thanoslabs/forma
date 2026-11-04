import React from "react";
import { createRoot } from "react-dom/client";
import "./style.css";
import { Office } from "./Office";

const sections = [
  { name: "Office", description: "The living workspace will appear here once the world renderer is connected." },
  { name: "Agents", description: "Connect a harness to discover agents and inspect their sessions." },
  { name: "Projects", description: "Connect Overwatch or another project provider to view work." },
];

function App() {
  const [selected, setSelected] = React.useState(0);
  return <div className="app">
    <aside className="sidebar"><h1>forma<span>.</span></h1><p className="eyebrow">WORKSPACE</p>
      <nav aria-label="Main navigation">{sections.map((section, i) =>
        <button key={section.name} className={selected === i ? "active" : ""} onClick={() => setSelected(i)}>{section.name}</button>)}</nav>
      <p className="footnote">Foundation preview · No providers connected</p>
    </aside>
    <main><header><span>Workspace / {sections[selected].name}</span><span className="status">● Setup required</span></header>
      <section className="content"><p className="eyebrow">WELCOME TO FORMA</p><h2>{sections[selected].name}</h2>
      <p>{sections[selected].description}</p>{selected === 0 ? <Office /> : <div className="placeholder" role="img" aria-label="Empty office placeholder">
      <div className="room"><span>✳</span><strong>Your workspace starts here</strong><small>Connect an agent harness and project provider to bring it to life.</small></div>
      </div>}</section>
    </main>
  </div>;
}

createRoot(document.getElementById("root")!).render(<React.StrictMode><App/></React.StrictMode>);
