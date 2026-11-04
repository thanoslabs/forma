import { describe,expect,it } from "vitest";
import { speechSegments } from "./speech";
import type { LinkedMention } from "./mentions";
const mention=(kind:"agent"|"project",text:string,start:number):LinkedMention=>({kind,id:text,name:text,text,start,end:start+text.length});
describe("speech mention highlights",()=>{
 it("highlights agent and project mentions",()=>{
  const body="Hi @Atlas on @project:Forma";
  const segments=speechSegments(body,[mention("agent","@Atlas",3),mention("project","@project:Forma",13)]);
  expect(segments.filter(s=>s.kind).map(s=>s.kind)).toEqual(["agent","project"]);
 });
 it("does not highlight unresolved text",()=>expect(speechSegments("hello @Unknown")).toEqual([{text:"hello @Unknown"}]));
 it("truncates at 120 characters without highlighting beyond cutoff",()=>{
  const body="a".repeat(130)+" @Atlas";
  expect(speechSegments(body,[mention("agent","@Atlas",131)]).some(s=>s.kind)).toBe(false);
  expect(Array.from(speechSegments(body).map(s=>s.text).join(""))).toHaveLength(120);
 });
 it("preserves mention color when cutoff occurs inside it",()=>{
  const body="a".repeat(115)+"@project:Forma"+" trailing";
  const segments=speechSegments(body,[mention("project","@project:Forma",115)]);
  expect(segments.some(s=>s.kind==="project")).toBe(true);
  expect(Array.from(segments.map(s=>s.text).join(""))).toHaveLength(120);
 });
});
