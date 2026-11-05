import { Container, Graphics, Text } from "pixi.js";
import type { Position } from "./Office";

/** Layered 2.5D office fixtures. All coordinates use the shared 860×490 world. */
export function makeWindows(): Container {
 const layer = new Container();
 for (const x of [255, 535]) {
  const sky = new Graphics()
   .roundRect(x, 18, 210, 104, 8).fill("#89c8ee")
   .rect(x + 3, 69, 204, 48).fill("#a5d7ee");
  // Layered clouds and distant silhouettes provide a sky texture without network assets.
  for (const [dx,dy,s] of [[28,29,1],[125,46,.8],[166,23,.55]] as const) {
   sky.ellipse(x+dx,18+dy,27*s,9*s).fill({color:"#f8fcff",alpha:.86});
   sky.ellipse(x+dx+17*s,18+dy-5*s,19*s,12*s).fill({color:"#ffffff",alpha:.82});
   sky.ellipse(x+dx-15*s,18+dy+2*s,16*s,7*s).fill({color:"#e6f6ff",alpha:.8});
  }
  sky.moveTo(x+3,115).lineTo(x+3,101).lineTo(x+43,87).lineTo(x+85,105).lineTo(x+130,90).lineTo(x+207,106).lineTo(x+207,115).closePath().fill("#9ab9b8");
  sky.roundRect(x,18,210,104,8).stroke({color:"#f9f8ef",width:8});
  sky.rect(x+102,20,6,100).fill("#f8f7f1");
  sky.rect(x+3,69,204,5).fill("#f8f7f1");
  sky.roundRect(x-6,122,222,8,2).fill("#bca58b");
  layer.addChild(sky);
 }
 return layer;
}
export function makeProjectBoard(onOpen: () => void): Container {
 const board = new Container();
 const art = new Graphics();
 art.ellipse(111,189,91,12).fill({color:"#625848",alpha:.18});
 art.roundRect(33,49,158,111,7).fill("#8a969e");
 art.roundRect(39,55,146,98,4).fill("#f9faf6").stroke({color:"#c5cdd1",width:3});
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
export function makeWorkstations(positions:Position[]): Container {
 const layer = new Container();
 positions.forEach((p,index)=>{
  const g=new Graphics();
  // Contact shadow, angled tabletop, thick front apron and four visible legs.
  g.ellipse(p.x,p.y+51,83,16).fill({color:"#4c3b32",alpha:.18});
  for(const dx of [-51,49]) g.roundRect(p.x+dx,p.y+25,8,51,2).fill("#6b493a");
  g.moveTo(p.x-70,p.y+2).lineTo(p.x+57,p.y+2).lineTo(p.x+72,p.y+27).lineTo(p.x-55,p.y+27).closePath().fill("#c9956d").stroke({color:"#9a6448",width:2});
  g.moveTo(p.x-55,p.y+27).lineTo(p.x+72,p.y+27).lineTo(p.x+70,p.y+38).lineTo(p.x-57,p.y+38).closePath().fill("#a46d51");
  if(index%2===0) {
   // Open laptop, raised screen, hinge and keyboard deck.
   g.roundRect(p.x-31,p.y-40,64,49,4).fill("#4c5965").stroke({color:"#29343e",width:3});
   g.roundRect(p.x-27,p.y-36,56,40,2).fill("#88b8d0");
   g.rect(p.x-24,p.y-31,49,6).fill("#d5e9f3");
   g.roundRect(p.x-35,p.y+8,74,14,3).fill("#66747c");
   for(let row=0;row<2;row++)for(let col=0;col<8;col++)g.roundRect(p.x-29+col*8,p.y+10+row*5,6,3,1).fill("#b6c2c9");
  } else {
   // Monitor with stand, separate keyboard and mouse.
   g.roundRect(p.x-37,p.y-51,78,51,4).fill("#343e4b").stroke({color:"#222b36",width:3});
   g.roundRect(p.x-32,p.y-46,68,40,2).fill("#83b3ce");
   g.roundRect(p.x-28,p.y-42,60,6,1).fill("#c6e2ed");
   g.rect(p.x-2,p.y,7,12).fill("#67717a");
   g.roundRect(p.x-14,p.y+11,32,5,2).fill("#67717a");
   g.roundRect(p.x-31,p.y+20,57,10,3).fill("#dce1df");
   for(let col=0;col<9;col++)g.rect(p.x-27+col*6,p.y+23,4,2).fill("#85949b");
   g.ellipse(p.x+43,p.y+24,7,10).fill("#e1e7e4");
   g.moveTo(p.x+43,p.y+16).lineTo(p.x+43,p.y+24).stroke({color:"#9daab0",width:1});
  }
  layer.addChild(g);
 });
 return layer;
}
