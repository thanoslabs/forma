import{afterEach,describe,expect,it,vi}from"vitest";
import{ensureDefaultConversation,listConversationMessages,sendConversationMessage}from"./conversationApi";

afterEach(()=>vi.unstubAllGlobals());
const response=(value:unknown,status=200)=>new Response(JSON.stringify(value),{status,headers:{"Content-Type":"application/json"}});
describe("persisted conversation API",()=>{
 it("bootstraps an empty workspace and conversation",async()=>{
  const fetch=vi.fn().mockResolvedValueOnce(response([])).mockResolvedValueOnce(response({id:"workspace-1",name:"Forma"},201)).mockResolvedValueOnce(response([])).mockResolvedValueOnce(response({id:"conversation-1",workspaceId:"workspace-1",title:"Office"},201));vi.stubGlobal("fetch",fetch);
  await expect(ensureDefaultConversation()).resolves.toEqual({workspaceId:"workspace-1",conversationId:"conversation-1"});
  expect(fetch).toHaveBeenCalledTimes(4);expect(JSON.parse(String((fetch.mock.calls[1][1] as RequestInit).body))).toEqual({name:"Forma"});
 });
 it("loads and sends workspace-scoped messages",async()=>{
  const fetch=vi.fn().mockResolvedValueOnce(response([])).mockResolvedValueOnce(response({id:"message-1"},201));vi.stubGlobal("fetch",fetch);const scope={workspaceId:"workspace-1",conversationId:"conversation-1"};
  await expect(listConversationMessages(scope)).resolves.toEqual([]);await sendConversationMessage(scope,"hello",[],"message-key");
  expect(fetch.mock.calls[0][0]).toContain("/workspaces/workspace-1/conversations/conversation-1/messages");expect(JSON.parse(String((fetch.mock.calls[1][1] as RequestInit).body))).toEqual({body:"hello",mentions:[]});expect((fetch.mock.calls[1][1] as RequestInit).headers).toMatchObject({"Idempotency-Key":"message-key"});
 });
 it("surfaces API errors",async()=>{vi.stubGlobal("fetch",vi.fn().mockResolvedValue(new Response("database unavailable",{status:503})));await expect(ensureDefaultConversation()).rejects.toThrow("database unavailable");});
});
