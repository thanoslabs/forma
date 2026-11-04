import type { OfficeLayout } from "./layout";
type Props={layout:OfficeLayout;onChange:(layout:OfficeLayout)=>void;onReset:()=>void};
export function LayoutEditor({layout,onChange,onReset}:Props){
 return <details className="layout-editor"><summary>Customize workstations</summary><p>Adjust desk positions. Layout saves locally in this browser.</p>
 {layout.desks.map((d,i)=><fieldset key={i}><legend>Desk {i+1}</legend>
  <label>X <input aria-label={`Desk ${i+1} horizontal position`} type="range" min="110" max="750" step="10" value={d.x} onChange={e=>onChange({...layout,desks:layout.desks.map((p,j)=>i===j?{...p,x:Number(e.target.value)}:p)})}/></label>
  <label>Y <input aria-label={`Desk ${i+1} vertical position`} type="range" min="130" max="290" step="10" value={d.y} onChange={e=>onChange({...layout,desks:layout.desks.map((p,j)=>i===j?{...p,y:Number(e.target.value)}:p)})}/></label>
 </fieldset>)}
 <button onClick={onReset}>Restore default layout</button></details>;
}
