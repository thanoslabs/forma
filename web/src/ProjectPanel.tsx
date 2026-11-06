import React from"react";
import{listProjects,listWorkItems,type ProjectProjection,type WorkItem}from"./projectApi";

export function ProjectPanel({workspaceId}:{workspaceId:string}){
 const[projects,setProjects]=React.useState<ProjectProjection[]>([]),[items,setItems]=React.useState<WorkItem[]>([]),[error,setError]=React.useState<string>(),[loading,setLoading]=React.useState(true);
 React.useEffect(()=>{const controller=new AbortController();setLoading(true);setError(undefined);void listProjects(workspaceId,controller.signal).then(async found=>{setProjects(found);const groups=await Promise.all(found.map(project=>listWorkItems(workspaceId,project,controller.signal)));setItems(groups.flat())}).catch(reason=>{if(!controller.signal.aborted)setError(reason instanceof Error?reason.message:"Unable to load projects")}).finally(()=>setLoading(false));return()=>controller.abort()},[workspaceId]);
 const names=new Map(projects.map(project=>[project.id,project.name]));
 return <div className="office-panel-section"><h3>Projects & work items</h3>{loading&&<p>Loading Overwatch…</p>}{error&&<p className="connection-error">{error}</p>}{!loading&&!error&&projects.length===0&&<p>Connect Overwatch in Settings to bring live work into this floor.</p>}{items.map(item=><article className="office-project" key={`${item.projectId}:${item.id}`}><div><small>{names.get(item.projectId)??"Project"}</small><strong>{item.title}</strong></div><span>{item.status}</span>{item.url&&<a href={item.url} target="_blank" rel="noreferrer">Open ↗</a>}</article>)}</div>;
}
