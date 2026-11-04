import type { SimState } from "./simulation";
type Props = { state: SimState; selectedId?: string; onSelect: (id: string) => void; onAssign: (taskId: string, agentId: string) => void };
export function OfficePanels({state,selectedId,onSelect,onAssign}:Props){
 const selected=state.agents.find(a=>a.id===selectedId);
 return <div className="office-panels">
  <section aria-label="Agent roster"><h3>Agents</h3>{state.agents.map(a=><button key={a.id} className={selectedId===a.id?"selected":""} onClick={()=>onSelect(a.id)}><strong>{a.name}</strong><span>{a.status} · {a.progress}%</span></button>)}
   {selected&&<div className="agent-detail"><h4>{selected.name}</h4><p>State: {selected.status}</p><p>Current task: {state.tasks.find(t=>t.id===selected.taskId)?.title??"None"}</p><progress value={selected.progress} max={100} aria-label="Task progress"/></div>}
  </section>
  <section aria-label="Task board"><h3>Task board</h3>{state.tasks.map(t=><div className="task" key={t.id}><strong>{t.title}</strong><span>{t.status}</span>{t.status==="queued"&&<select aria-label={`Assign ${t.title}`} value="" onChange={e=>{if(e.target.value)onAssign(t.id,e.target.value);}}><option value="">Assign to…</option>{state.agents.filter(a=>a.status==="idle"||a.status==="completed").map(a=><option key={a.id} value={a.id}>{a.name}</option>)}</select>}</div>)}</section>
  <section aria-label="Activity feed"><h3>Activity</h3><ol>{state.events.map(e=><li key={e.id}><small>Tick {e.tick}</small> {e.message}</li>)}</ol></section>
 </div>;
}
