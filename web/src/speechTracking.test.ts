import {describe,expect,it} from "vitest";
import {speakerPosition,type SpeechBubble} from "./speech";
const bubble:SpeechBubble={id:"1",speakerId:"atlas",speakerName:"Atlas",body:"Hello",messageId:"1",position:{x:100,y:100}};
describe("speech bubble tracking",()=>{
 it("follows the latest agent position",()=>expect(speakerPosition(bubble,{x:1,y:2},[{id:"atlas",position:{x:300,y:280}}])).toEqual({x:300,y:280}));
 it("follows the player when the player speaks",()=>expect(speakerPosition({...bubble,speakerId:"player"},{x:400,y:350},[])).toEqual({x:400,y:350}));
 it("retains the last known position if an agent disappears",()=>expect(speakerPosition(bubble,{x:1,y:2},[])).toEqual({x:100,y:100}));
});
