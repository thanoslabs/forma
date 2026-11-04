export type MentionKind = "agent" | "project";
export type MentionTarget = { kind:MentionKind; id:string; name:string };
export type LinkedMention = MentionTarget & { start:number; end:number; text:string };
export type MentionMatch = { start:number; end:number; query:string; kind?:MentionKind };
const isWord=(c:string)=>/[\\w]/.test(c);
export function activeMention(text:string,caret:number):MentionMatch|null {
 const before=text.slice(0,caret);
 const match=/(^|[\\s(])@(agent:|project:)?([^\\s@]*)$/i.exec(before);
 if(!match)return null;
 const start=caret-match[0].length+match[1].length;
 return {start,end:caret,kind:match[2]?match[2].slice(0,-1).toLowerCase() as MentionKind:undefined,query:match[3]};
}
export function mentionOptions(match:MentionMatch,targets:MentionTarget[]):MentionTarget[] {
 const q=match.query.toLocaleLowerCase();
 return targets.filter(t=>(!match.kind||match.kind===t.kind)&&(t.name.toLocaleLowerCase().includes(q)||t.id.toLocaleLowerCase().includes(q)))
  .sort((a,b)=>Number(b.name.toLowerCase().startsWith(q))-Number(a.name.toLowerCase().startsWith(q))||a.name.localeCompare(b.name)).slice(0,8);
}
export function mentionText(target:MentionTarget,targets:MentionTarget[]):string {
 const ambiguous=targets.some(t=>t.kind!==target.kind&&t.name.toLowerCase()===target.name.toLowerCase());
 return "@"+(ambiguous?target.kind+":":"")+target.name.replace(/\\s+/g,"-");
}
export function resolveMentions(text:string,targets:MentionTarget[]):LinkedMention[] {
 const results:LinkedMention[]=[];
 const re=/(^|[\\s(])@(agent:|project:)?([\\w-]+)/gi;
 let m:RegExpExecArray|null;
 while((m=re.exec(text))!==null){
  const kind=m[2]?.slice(0,-1).toLowerCase() as MentionKind|undefined;
  const value=m[3].toLowerCase(),matches=targets.filter(t=>(!kind||t.kind===kind)&&(t.name.toLowerCase().replace(/\\s+/g,"-")===value||t.id.toLowerCase()===value));
  if(matches.length!==1)continue; // Ambiguous unqualified names must be disambiguated.
  const start=m.index+m[1].length;
  results.push({...matches[0],start,end:start+m[0].length-m[1].length,text:text.slice(start,start+m[0].length-m[1].length)});
 }
 return results;
}
export function insertMention(text:string,match:MentionMatch,target:MentionTarget,all:MentionTarget[]):string { return text.slice(0,match.start)+mentionText(target,all)+" "+text.slice(match.end); }
