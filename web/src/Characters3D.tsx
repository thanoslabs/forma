import { useEffect, useRef } from "react";
import * as THREE from "three";
import type { OfficeAgent } from "./Office";

/** All dimensions are in world units. Characters are original procedural 3D geometry. */
export type CharacterStyle = "agent" | "explorer";
export type CharacterMotion = "idle" | "walk" | "work" | "sit" | "wave";
export const PLAYER_PRESETS = ["explorer", "builder", "creative", "captain", "researcher", "cozy"] as const;
export const DEFAULT_PLAYER_PRESET = "explorer";

type Puppet = { root: THREE.Group; leftArm: THREE.Group; rightArm: THREE.Group; leftLeg: THREE.Group; rightLeg: THREE.Group; head: THREE.Group };
const mat = (color: number, roughness = 0.75) => new THREE.MeshStandardMaterial({ color, roughness });
const mesh = (geometry: THREE.BufferGeometry, material: THREE.Material, parent: THREE.Object3D, x: number, y: number, z: number) => {
 const item = new THREE.Mesh(geometry, material); item.position.set(x,y,z); parent.add(item); return item;
};
const ball = (parent: THREE.Object3D, material: THREE.Material, x: number,y: number,z: number,sx: number,sy: number,sz: number) => {
 const m = mesh(new THREE.SphereGeometry(1,20,14),material,parent,x,y,z); m.scale.set(sx,sy,sz); return m;
};
const roundedBox = (parent: THREE.Object3D, material: THREE.Material, x: number,y: number,z: number,w: number,h: number,d: number,r: number) => {
 const shape = new THREE.Shape();
 const a=-w/2,b=-h/2;
 shape.moveTo(a+r,b);shape.lineTo(a+w-r,b);shape.quadraticCurveTo(a+w,b,a+w,b+r);
 shape.lineTo(a+w,b+h-r);shape.quadraticCurveTo(a+w,b+h,a+w-r,b+h);
 shape.lineTo(a+r,b+h);shape.quadraticCurveTo(a,b+h,a,b+h-r);
 shape.lineTo(a,b+r);shape.quadraticCurveTo(a,b,a+r,b);
 const geo = new THREE.ExtrudeGeometry(shape,{depth:d-2*r,bevelEnabled:true,bevelThickness:r,bevelSize:r,bevelSegments:3,steps:1,curveSegments:4});
 geo.center();return mesh(geo,material,parent,x,y,z);
};
export function createCharacter(style: CharacterStyle, accent = 0x9271d6): Puppet {
 const root = new THREE.Group(), head = new THREE.Group();
 const leftArm = new THREE.Group(),rightArm = new THREE.Group(),leftLeg = new THREE.Group(),rightLeg = new THREE.Group();
 const dark=mat(0x24223c), shoe=mat(0x54443f), skin=mat(0xf4bd91), cyan=new THREE.MeshStandardMaterial({color:0x72e7ff,emissive:0x1d849b,emissiveIntensity:0.45});
 root.add(head,leftArm,rightArm,leftLeg,rightLeg);
 head.position.y=1.32;
 leftArm.position.set(-0.56,1.04,0);rightArm.position.set(0.56,1.04,0);
 leftLeg.position.set(-0.27,0.53,0);rightLeg.position.set(0.27,0.53,0);
 if(style==="agent"){
  const shell=mat(accent),trim=mat(0xb4a6dd),face=mat(0x15162b,0.24);
  roundedBox(root,shell,0,0.92,0,1.02,0.79,0.62,0.17);
  roundedBox(root,cyan,0,0.97,0.35,0.25,0.18,0.07,0.03);
  roundedBox(head,shell,0,0.18,0,1.48,1.08,0.7,0.24);
  roundedBox(head,face,0,0.18,0.382,1.15,0.76,0.055,0.15);
  for(const x of [-0.28,0.28]) {
   const eye=ball(head,cyan,x,0.22,0.435,0.12,0.075,0.022);eye.rotation.z=x<0?0.17:-0.17;
  }
  ball(head,trim,0,0.85,0,0.075,0.16,0.075);
  ball(head,cyan,0,1.02,0,0.12,0.12,0.12);
  for(const side of [-1,1]) {
   const arm=side<0?leftArm:rightArm,leg=side<0?leftLeg:rightLeg;
   ball(head,trim,side*0.79,0.18,0,0.19,0.22,0.2);
   ball(arm,shell,side*0.05,-0.24,0,0.23,0.31,0.23);
   ball(arm,dark,side*0.05,-0.48,0.03,0.2,0.17,0.21);
   roundedBox(leg,shell,0,-0.19,0,0.41,0.43,0.4,0.11);
   ball(leg,dark,0,-0.4,0.13,0.24,0.12,0.31);
  }
 }else{
  const hoodie=mat(0x279a94),cuff=mat(0x1a7978),hair=mat(0x49312e),beanie=mat(0xf2b43e),white=mat(0xf7e8d5);
  ball(root,hoodie,0,0.94,0,0.62,0.49,0.38);
  roundedBox(root,hoodie,0,0.82,0.13,0.92,0.62,0.55,0.19);
  for(const x of [-0.12,0.12]) roundedBox(root,white,x,1.05,0.42,0.035,0.29,0.03,0.01);
  ball(head,skin,0,0.11,0.06,0.7,0.59,0.55);
  ball(head,hair,0,0.48,-0.04,0.72,0.38,0.56);
  for(const x of [-0.42,-0.16,0.12,0.38]) ball(head,hair,x,0.39,0.44,0.22,0.22,0.16);
  ball(head,beanie,0,0.68,0,0.72,0.3,0.57);
  ball(head,beanie,0,0.54,0.08,0.78,0.13,0.62);
  ball(head,beanie,0,0.96,0,0.17,0.12,0.16);
  for(const x of [-0.25,0.25]) ball(head,dark,x,0.12,0.571,0.045,0.095,0.025);
  ball(head,mat(0xb77960),0,-0.08,0.59,0.055,0.025,0.015);
  for(const side of [-1,1]){
   const arm=side<0?leftArm:rightArm,leg=side<0?leftLeg:rightLeg;
   ball(arm,hoodie,0,-0.23,0,0.25,0.32,0.27);
   ball(arm,cuff,0,-0.46,0,0.22,0.12,0.23);
   ball(arm,skin,0,-0.56,0,0.18,0.13,0.19);
   roundedBox(leg,mat(0x675c58),0,-0.18,0,0.39,0.42,0.38,0.12);
   ball(leg,shoe,0,-0.42,0.14,0.25,0.14,0.36);
  }
 }
 return {root,head,leftArm,rightArm,leftLeg,rightLeg};
}
export function poseCharacter(p:Puppet, motion:CharacterMotion, time:number) {
 const walk=motion==="walk",work=motion==="work",sit=motion==="sit",wave=motion==="wave";
 const phase=time*7;
 p.root.position.y=walk?Math.abs(Math.sin(phase))*0.07:Math.sin(time*2)*0.012;
 p.leftLeg.rotation.x=walk?Math.sin(phase)*0.55:sit?-1.15:0;
 p.rightLeg.rotation.x=walk?-Math.sin(phase)*0.55:sit?-1.15:0;
 p.leftArm.rotation.x=walk?-Math.sin(phase)*0.55:work?-0.65+Math.sin(phase)*0.13:0;
 p.rightArm.rotation.x=walk?Math.sin(phase)*0.55:work?-0.65-Math.sin(phase)*0.13:wave?-2.2+Math.sin(time*8)*0.3:0;
 p.rightArm.rotation.z=wave?-0.45:0;
 p.head.rotation.z=Math.sin(time*1.4)*0.025;
}
export function motionForAgent(agent:OfficeAgent, moving:boolean):CharacterMotion {
 if(moving)return "walk";
 if(agent.status==="working")return "work";
 return "idle";
}

/** Transparent Three.js layer aligned with the 860×490 Pixi office. Pointer events remain with Pixi. */
export function OfficeCharacters({agents}:{agents:OfficeAgent[]}) {
 const host=useRef<HTMLDivElement>(null), latest=useRef(agents);latest.current=agents;
 useEffect(()=>{
  const el=host.current;if(!el)return;
  const scene=new THREE.Scene(),camera=new THREE.OrthographicCamera(-8.6,8.6,4.9,-4.9,0.1,100);
  camera.position.set(0,0,25);camera.lookAt(0,0,0);
  scene.add(new THREE.HemisphereLight(0xffffff,0x897da0,2.2));
  const light=new THREE.DirectionalLight(0xffffff,2.2);light.position.set(-3,7,9);scene.add(light);
  const renderer=new THREE.WebGLRenderer({alpha:true,antialias:true});
  renderer.setPixelRatio(Math.min(window.devicePixelRatio||1,2));renderer.setSize(860,490);
  renderer.outputColorSpace=THREE.SRGBColorSpace;renderer.domElement.style.width="100%";renderer.domElement.style.height="100%";el.appendChild(renderer.domElement);
  const models=new Map<string,{puppet:Puppet,x:number,y:number,initialized:boolean}>();
  const player=createCharacter("explorer");scene.add(player.root);player.root.scale.setScalar(0.46);player.root.position.set(-6.4,-3.4,0);
  const clock=new THREE.Clock();let frame=0;
  const animate=()=>{
   frame=requestAnimationFrame(animate);
   const t=clock.getElapsedTime(),ids=new Set(latest.current.map(a=>a.id));
   for(const [id,entry] of models)if(!ids.has(id)){scene.remove(entry.puppet.root);models.delete(id);}
   for(const agent of latest.current){
    let entry=models.get(agent.id);
    if(!entry){const puppet=createCharacter("agent",agent.color??0x9271d6);scene.add(puppet.root);entry={puppet,x:agent.position.x,y:agent.position.y,initialized:true};models.set(agent.id,entry);}
    const dx=agent.position.x-entry.x,dy=agent.position.y-entry.y;
    const moving=Math.hypot(dx,dy)>2;
    entry.x+=dx*0.12;entry.y+=dy*0.12;
    entry.puppet.root.position.x=(entry.x-430)/50;
    entry.puppet.root.position.z=0;
    entry.puppet.root.position.y=(245-entry.y)/50-0.55;
    entry.puppet.root.scale.setScalar(0.46);
    poseCharacter(entry.puppet,motionForAgent(agent,moving),t);
   }
   poseCharacter(player,"idle",t);
   renderer.render(scene,camera);
  };animate();
  return()=>{cancelAnimationFrame(frame);models.forEach(({puppet})=>{puppet.root.traverse(o=>{if(o instanceof THREE.Mesh){o.geometry.dispose();const m=o.material;if(Array.isArray(m))m.forEach(v=>v.dispose());else m.dispose();}});});renderer.dispose();renderer.domElement.remove();};
 },[]);
 return <div ref={host} className="office-characters" aria-hidden="true"/>;
}
