import type { Position } from "./Office";
import type { LinkedMention } from "./mentions";
export type SpeechBubble = { id:string; speakerId:string; speakerName:string; body:string; mentions?:LinkedMention[]; position:Position; messageId:string };
export function speakerPosition(bubble:SpeechBubble,player:Position,agents:ReadonlyArray<{id:string;position:Position}>):Position {
 if(bubble.speakerId==="player")return player;
 return agents.find(agent=>agent.id===bubble.speakerId)?.position??bubble.position;
}
export const SPEECH_LIMIT=120;
export function speechPreview(body:string):string {
 const chars=Array.from(body);
 return chars.length<=SPEECH_LIMIT?body:chars.slice(0,SPEECH_LIMIT-1).join("")+"…";
}
export function visibleBubbles(bubbles:SpeechBubble[]):SpeechBubble[] {
 // Only the newest message for each speaker remains visible.
 const seen=new Set<string>();
 return [...bubbles].reverse().filter(b=>{if(seen.has(b.speakerId))return false;seen.add(b.speakerId);return true;}).reverse();
}

export type SpeechSegment={text:string;kind?:"agent"|"project"};
export function speechSegments(body:string,mentions:LinkedMention[]=[]):SpeechSegment[] {
 const preview=speechPreview(body);
 const end=preview.endsWith("…")&&Array.from(body).length>SPEECH_LIMIT?Array.from(body).slice(0,SPEECH_LIMIT-1).join("").length:body.length;
 const result:SpeechSegment[]=[];
 let cursor=0;
 for(const mention of [...mentions].sort((a,b)=>a.start-b.start)){
  if(mention.start<cursor||mention.start>=end)continue;
  if(mention.start>cursor)result.push({text:body.slice(cursor,mention.start)});
  const stop=Math.min(mention.end,end);
  result.push({text:body.slice(mention.start,stop),kind:mention.kind});
  cursor=stop;
 }
 if(cursor<end)result.push({text:body.slice(cursor,end)});
 if(end<body.length)result.push({text:"…"});
 return result;
}
