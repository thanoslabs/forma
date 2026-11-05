import { Container, Graphics, Text } from "pixi.js";
import type { Position } from "./Office";

/** Layered 2.5D office fixtures. All coordinates use the shared 860×490 world. */
export function makeWindows(): Container {
 const layer = new Container();
 for (const x of [255, 535]) {
  const sky = new Graphics()
   .rect(x, 18, 210, 104).fill("#69a9d2")
   .rect(x + 3, 69, 204, 48).fill("#a5d7ee");
  // Layered clouds and distant silhouettes provide a sky texture without network assets.
  for (const [dx,dy,s] of [[28,29,1],[125,46,.8],[166,23,.55]] as const) {
   sky.ellipse(x+dx,18+dy,27*s,9*s).fill({color:"#f8fcff",alpha:.86});
   sky.ellipse(x+dx+17*s,18+dy-5*s,19*s,12*s).fill({color:"#ffffff",alpha:.82});
   sky.ellipse(x+dx-15*s,18+dy+2*s,16*s,7*s).fill({color:"#e6f6ff",alpha:.8});
  }
  sky.moveTo(x+3,115).lineTo(x+3,101).lineTo(x+43,87).lineTo(x+85,105).lineTo(x+130,90).lineTo(x+207,106).lineTo(x+207,115).closePath().fill("#9ab9b8");
  sky.rect(x,18,210,104).stroke({color:"#423b5c",width:8});
  sky.rect(x+102,20,6,100).fill("#f8f7f1");
  sky.rect(x+3,69,204,5).fill("#f8f7f1");
  sky.rect(x-6,122,222,8).fill("#796584");
  layer.addChild(sky);
 }
 return layer;
}
export function makeProjectBoard(onOpen: () => void): Container {
 const board = new Container();
 const art = new Graphics();
 art.ellipse(111,189,91,12).fill({color:"#625848",alpha:.18});
 art.rect(33,49,158,111).fill("#554c70");
 art.rect(39,55,146,98).fill("#f7f1df").stroke({color:"#302b49",width:3});
 art.rect(48,65,128,2).fill("#dce3e7");
 art.roundRect(54,81,34,20,3).fill("#d4e4fa");
 art.roundRect(94,81,34,20,3).fill("#e7dcf7");
 art.roundRect(134,81,34,20,3).fill("#d9f0df");
 for(const x of [54,94,134]) {art.roundRect(x,108,30,5,2).fill("#b5bec9");art.roundRect(x,118,22,4,2).fill("#d3d9df");}
 art.rect(106,160,7,24).fill("#77828c");
 art.moveTo(76,183).lineTo(143,183).stroke({color:"#77828c",width:7});
 for(const x of [77,143]) {art.circle(x,188,7).fill("#303b46");art.circle(x,188,3).fill("#8b9aa4");}
 art.roundRect(85,153,54,6,2).fill("#b7c1c7");
 board.addChild(art);
 const title = new Text({text:"PROJECTS",style:{fontFamily:"sans-serif",fontSize:13,fontWeight:"bold",fill:"#34465d"}});
 title.position.set(80,61);title.anchor.x=.5;board.addChild(title);
 const hit = new Graphics().roundRect(31,47,162,143,7).fill({color:0xffffff,alpha:.001});
 hit.eventMode="static";hit.cursor="pointer";hit.on("pointertap",e=>{e.stopPropagation();onOpen();});
 board.addChild(hit);
 return board;
}
export function makeElevator(onOpen:()=>void):Container{
 const elevator=new Container(),art=new Graphics();art.roundRect(772,105,74,172,8).fill("#596775").stroke({color:"#34414f",width:4});art.roundRect(780,115,58,151,4).fill("#b9c2c9");art.rect(808,115,3,151).fill("#6f7d88");art.roundRect(747,151,20,36,4).fill("#303c49");art.circle(757,162,4).fill("#a6e5c8");art.circle(757,177,4).fill("#e4bd73");elevator.addChild(art);const label=new Text({text:"ELEVATOR",style:{fontFamily:"sans-serif",fontSize:10,fontWeight:"bold",fill:"#eaf0f3"}});label.anchor.set(.5);label.position.set(809,96);elevator.addChild(label);const hit=new Graphics().roundRect(744,94,104,187,8).fill({color:0xffffff,alpha:.001});hit.eventMode="static";hit.cursor="pointer";hit.on("pointertap",event=>{event.stopPropagation();onOpen();});elevator.addChild(hit);return elevator;
}
export function makeWorkstations(positions:Position[]): Container[] {
 const layers:Container[]=[];
 positions.forEach((p,index)=>{
  const back=new Container(),front=new Container(),g=new Graphics(),fg=new Graphics();back.zIndex=p.y;front.zIndex=p.y+42;
  // Furniture is calibrated against a roughly 92px character: 70px top, 45px floor height.
  g.ellipse(p.x,p.y+52,68,13).fill({color:"#4c3b32",alpha:.16});
  // Chair is a separate interaction object behind the desk foreground.
  g.rect(p.x-21,p.y+42,42,34).fill("#59627b").stroke({color:"#302f49",width:3});
  g.rect(p.x-24,p.y+67,48,12).fill("#444a68");g.rect(p.x-3,p.y+78,6,18).fill("#343851");
  for(const dx of [-16,16])g.circle(p.x+dx,p.y+97,5).fill("#293545");
  for(const dx of [-43,43]) fg.rect(p.x+dx,p.y+23,8,43).fill("#694734");
  g.moveTo(p.x-58,p.y+2).lineTo(p.x+48,p.y+2).lineTo(p.x+59,p.y+22).lineTo(p.x-47,p.y+22).closePath().fill("#c9956d").stroke({color:"#9a6448",width:2});
  fg.moveTo(p.x-47,p.y+22).lineTo(p.x+59,p.y+22).lineTo(p.x+57,p.y+31).lineTo(p.x-49,p.y+31).closePath().fill("#a46d51");
  if(index%2===0) {
   // Open laptop, raised screen, hinge and keyboard deck.
   g.roundRect(p.x-27,p.y-34,56,40,4).fill("#4c5965").stroke({color:"#29343e",width:3});g.roundRect(p.x-23,p.y-30,48,32,2).fill("#88b8d0");g.rect(p.x-20,p.y-26,41,5).fill("#d5e9f3");g.roundRect(p.x-31,p.y+7,64,11,3).fill("#66747c");
  } else {
   // Monitor with stand, separate keyboard and mouse.
   g.roundRect(p.x-31,p.y-43,64,43,4).fill("#343e4b").stroke({color:"#222b36",width:3});g.roundRect(p.x-27,p.y-39,56,34,2).fill("#83b3ce");g.roundRect(p.x-23,p.y-35,48,5,1).fill("#c6e2ed");
   g.rect(p.x-2,p.y,7,12).fill("#67717a");
   g.roundRect(p.x-14,p.y+11,32,5,2).fill("#67717a");
   g.roundRect(p.x-31,p.y+20,57,10,3).fill("#dce1df");
   for(let col=0;col<9;col++)g.rect(p.x-27+col*6,p.y+23,4,2).fill("#85949b");
   g.ellipse(p.x+43,p.y+24,7,10).fill("#e1e7e4");
   g.moveTo(p.x+43,p.y+16).lineTo(p.x+43,p.y+24).stroke({color:"#9daab0",width:1});
  }
  back.addChild(g);front.addChild(fg);layers.push(back,front);
 });
 return layers;
}
