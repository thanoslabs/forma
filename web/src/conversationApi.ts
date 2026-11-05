import type { LinkedMention } from "./mentions";

export type ConversationMessage = {
 id:string;
 body:string;
 author:{kind:"user"|"agent"|"system";id:string;name:string};
 mentions:LinkedMention[];
 occurredAt:string;
 replyToMessageId?:string;
};

type SendResponse={messages:ConversationMessage[]};
const endpoint="/api/v1/demo/conversations/default/messages";

export async function listConversationMessages(signal?:AbortSignal):Promise<ConversationMessage[]> {
 const response=await fetch(endpoint,{signal});
 if(!response.ok)throw new Error(await responseText(response,"Unable to load the conversation"));
 return response.json() as Promise<ConversationMessage[]>;
}

export async function sendConversationMessage(body:string,mentions:LinkedMention[],idempotencyKey:string):Promise<ConversationMessage[]> {
 const response=await fetch(endpoint,{method:"POST",headers:{"Content-Type":"application/json","Idempotency-Key":idempotencyKey},body:JSON.stringify({body,mentions})});
 if(!response.ok)throw new Error(await responseText(response,"Unable to send the message"));
 return ((await response.json()) as SendResponse).messages;
}

async function responseText(response:Response,fallback:string):Promise<string>{const message=(await response.text()).trim();return message||fallback;}
