import { describe, expect, it } from "vitest";
import { activeMention, insertMention, mentionOptions, resolveMentions, type MentionTarget } from "./mentions";
const targets:MentionTarget[]=[
 {kind:"agent",id:"atlas",name:"Atlas"},
 {kind:"agent",id:"nova",name:"Nova"},
 {kind:"project",id:"atlas-project",name:"Atlas"},
 {kind:"project",id:"office-redesign",name:"Office Redesign"},
 {kind:"project",id:"forma",name:"Forma"}
];
describe("chat mentions",()=>{
 it("finds an active mention and typed prefix",()=>{expect(activeMention("hello @agent:At",15)).toEqual({start:6,end:15,kind:"agent",query:"At"});});
 it("filters by explicit kind",()=>{expect(mentionOptions(activeMention("@project:at",11)!,targets).map(t=>t.kind)).toEqual(["project"]);});
 it("suggests both kinds for an ambiguous name",()=>{expect(mentionOptions(activeMention("@Atlas",6)!,targets)).toHaveLength(2);});
 it("inserts qualified mention for ambiguous name",()=>{expect(insertMention("hi @At",activeMention("hi @At",6)!,targets[0],targets)).toBe("hi @agent:Atlas ");});
 it("resolves explicit agent and project identifiers",()=>{expect(resolveMentions("ask @agent:Atlas about @project:Office-Redesign",targets).map(m=>m.kind)).toEqual(["agent","project"]);});
 it("does not link ambiguous bare mentions",()=>{expect(resolveMentions("ask @Atlas",targets)).toHaveLength(0);});
 it("links unique bare mentions",()=>{expect(resolveMentions("hi @Nova and @Forma",targets).map(m=>m.id)).toEqual(["nova","forma"]);});
 it("ignores email addresses",()=>{expect(resolveMentions("hello@example.com",targets)).toHaveLength(0);});
});
