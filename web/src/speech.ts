import type { Position } from "./Office";
export type SpeechBubble = { id:string; speakerId:string; speakerName:string; body:string; position:Position; messageId:string };
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
