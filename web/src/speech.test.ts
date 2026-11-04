import { describe,expect,it } from "vitest";
import { speechPreview, visibleBubbles, SPEECH_LIMIT, type SpeechBubble } from "./speech";
describe("speech bubbles",()=>{
 it("preserves short messages",()=>expect(speechPreview("hello")).toBe("hello"));
 it("truncates long messages to exactly 120 characters",()=>{
  const result=speechPreview("x".repeat(180));expect(Array.from(result)).toHaveLength(SPEECH_LIMIT);expect(result.endsWith("…")).toBe(true);
 });
 it("handles unicode characters",()=>expect(Array.from(speechPreview("🙂".repeat(130)))).toHaveLength(120));
 it("shows only newest bubble per speaker",()=>{
  const base={speakerName:"Atlas",position:{x:200,y:200},messageId:"1",body:"hi"};
  const bubbles:SpeechBubble[]=[{...base,id:"1",speakerId:"atlas"},{...base,id:"2",speakerId:"player"},{...base,id:"3",speakerId:"atlas"}];
  expect(visibleBubbles(bubbles).map(b=>b.id)).toEqual(["2","3"]);
 });
});
