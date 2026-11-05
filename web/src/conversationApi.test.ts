import {afterEach,describe,expect,it,vi} from "vitest";
import {listConversationMessages,sendConversationMessage} from "./conversationApi";

afterEach(()=>vi.unstubAllGlobals());
describe("conversation API",()=>{
 it("loads the default conversation",async()=>{
  const fetch=vi.fn().mockResolvedValue(new Response("[]",{status:200,headers:{"Content-Type":"application/json"}}));vi.stubGlobal("fetch",fetch);
  await expect(listConversationMessages()).resolves.toEqual([]);
  expect(fetch).toHaveBeenCalledWith("/api/v1/demo/conversations/default/messages",{signal:undefined});
 });
 it("sends mentions with an idempotency key",async()=>{
  const fetch=vi.fn().mockResolvedValue(new Response('{"messages":[]}',{status:200,headers:{"Content-Type":"application/json"}}));vi.stubGlobal("fetch",fetch);
  const mentions=[{kind:"agent" as const,id:"atlas",name:"Atlas",text:"@Atlas",start:6,end:12}];
  await sendConversationMessage("hello @Atlas",mentions,"request-1");
  const init=fetch.mock.calls[0][1] as RequestInit;
  expect(init.method).toBe("POST");
  expect(init.headers).toEqual({"Content-Type":"application/json","Idempotency-Key":"request-1"});
  expect(JSON.parse(String(init.body))).toEqual({body:"hello @Atlas",mentions});
 });
 it("surfaces an API error",async()=>{
  vi.stubGlobal("fetch",vi.fn().mockResolvedValue(new Response("agent not found",{status:422})));
  await expect(sendConversationMessage("hello",[],"request-2")).rejects.toThrow("agent not found");
 });
});
