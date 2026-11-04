import type { SimAgent } from "./simulation";
export const palette = [0x7e70cf,0xe58d9a,0x6abbb1,0xe6b76b,0x7b9edc,0xb38c70];
type Props={agents:SimAgent[];colors:Record<string,number>;onChange:(id:string,color:number)=>void};
export function AvatarCustomizer({agents,colors,onChange}:Props){
 return <details className="layout-editor"><summary>Customize avatars</summary><p>Choose a color for each simulated agent.</p>
 {agents.map(a=><fieldset key={a.id}><legend>{a.name}</legend>{palette.map(color=><button key={color} type="button" aria-label={`Set ${a.name} color to #${color.toString(16)}`} aria-pressed={(colors[a.id]??a.color)===color} onClick={()=>onChange(a.id,color)} style={{background:`#${color.toString(16).padStart(6,"0")}`,width:32,height:32,border:(colors[a.id]??a.color)===color?"3px solid white":"1px solid #666"}}/>)}</fieldset>)}
 </details>;
}
