import { useEffect, useState } from "react";

type Agent = { id: string; name: string; status: string; lastSeen: string };
type Project = { id: string; name: string };
type Item = { id: string; title: string; status: string };
type Props = { kind: "agents" | "projects" };

export function DemoData({ kind }: Props) {
 const [agents, setAgents] = useState<Agent[]>([]);
 const [projects, setProjects] = useState<Project[]>([]);
 const [items, setItems] = useState<Record<string, Item[]>>({});
 const [error, setError] = useState<string | null>(null);
 useEffect(() => {
  const controller = new AbortController();
  const run = async () => {
   try {
    const response = await fetch(`/api/v1/demo/${kind}`, { signal: controller.signal });
    if (!response.ok) throw new Error("Demo API is unavailable. Start Forma with FORMA_ENABLE_DEMO_API=true.");
    if (kind === "agents") setAgents(await response.json() as Agent[]);
    else {
     const found = await response.json() as Project[];
     setProjects(found);
     const loaded = await Promise.all(found.map(async p => {
      const r = await fetch(`/api/v1/demo/projects/${encodeURIComponent(p.id)}/items`, { signal: controller.signal });
      if (!r.ok) throw new Error("Unable to fetch project work items");
      return [p.id, await r.json() as Item[]] as const;
     }));
     setItems(Object.fromEntries(loaded));
    }
    setError(null);
   } catch (e) { if (!controller.signal.aborted) setError(e instanceof Error ? e.message : "Unable to load demo"); }
  };
  void run();
  return () => controller.abort();
 }, [kind]);
 return <div className="demo-list" aria-live="polite">
  {error && <p role="status">{error}</p>}
  {kind === "agents" ? agents.map(a => <article key={a.id}><strong>{a.name}</strong><span>{a.status}</span></article>)
   : projects.map(p => <article key={p.id}><strong>{p.name}</strong>{(items[p.id] ?? []).map(i => <p key={i.id}>{i.title} · {i.status}</p>)}</article>)}
 </div>;
}
