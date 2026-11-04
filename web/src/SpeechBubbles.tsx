import { speakerPosition, speechSegments, visibleBubbles, type SpeechBubble } from "./speech";
import type { Position, OfficeAgent } from "./Office";
import "./speech.css";
export function SpeechBubbles({bubbles,playerPosition,agents,onOpen}:{bubbles:SpeechBubble[];playerPosition:Position;agents:OfficeAgent[];onOpen:(messageId:string)=>void}){
 return <div className="speech-layer" aria-label="Office speech bubbles">{visibleBubbles(bubbles).map(b=>
  <button type="button" key={b.id} className={"speech-bubble "+(b.speakerId==="player"?"speech-player":"")} style={{left:(speakerPosition(b,playerPosition,agents).x/860*100)+"%",top:((speakerPosition(b,playerPosition,agents).y-64)/490*100)+"%"}} onClick={()=>onOpen(b.messageId)} title={"Open "+b.speakerName+"'s full message in chat"}>
   <strong>{b.speakerName}</strong><span>{speechSegments(b.body,b.mentions).map((segment,i)=>segment.kind?<mark key={i} className={"speech-mention "+segment.kind}>{segment.text}</mark>:<span key={i}>{segment.text}</span>)}</span>
  </button>
 )}</div>;
}
