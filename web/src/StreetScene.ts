import{Assets,Container,Graphics,Sprite,Text}from"pixi.js";
import{publicAsset}from"./publicAsset";

export async function makeStreetScene(onEnterBuilding:()=>void){
 const layer=new Container(),texture=await Assets.load(publicAsset("assets/scenes/street.png")),background=new Sprite(texture);background.width=860;background.height=490;layer.addChild(background);
 const entrance=new Graphics().roundRect(353,112,150,174,10).fill({color:0xffffff,alpha:.001});entrance.eventMode="static";entrance.cursor="pointer";entrance.on("pointertap",event=>{event.stopPropagation();onEnterBuilding();});layer.addChild(entrance);
 const label=new Text({text:"ENTER BUILDING",style:{fontFamily:"sans-serif",fontSize:11,fontWeight:"bold",fill:0xffffff,stroke:{color:0x263548,width:4}}});label.anchor.set(.5);label.position.set(428,294);layer.addChild(label);
 const traffic=new Container();traffic.sortableChildren=true;layer.addChild(traffic);
 const colors=[0xe35f55,0x4d88c7,0xe2ad48,0x65a879];
 const cars=colors.map((color,index)=>{const car=new Container(),body=new Graphics();body.roundRect(-26,-11,52,22,7).fill(color).stroke({color:0x26313e,width:2});body.roundRect(-13,-14,27,13,5).fill(0xb9def0);for(const x of [-17,17])body.circle(x,11,5).fill(0x27313d);car.addChild(body);car.y=index%2?346:306;car.x=index*235;car.zIndex=car.y;traffic.addChild(car);return{car,speed:index%2?-.065:.075};});
 const signal=new Graphics();signal.roundRect(815,274,22,45,6).fill(0x27313d);signal.circle(826,287,6).fill(0xe65d54);signal.circle(826,304,6).fill(0x60bd75);signal.zIndex=400;traffic.addChild(signal);
 return{layer,update:(deltaMS:number)=>{for(const entry of cars){entry.car.x+=entry.speed*deltaMS;if(entry.speed>0&&entry.car.x>900)entry.car.x=-60;if(entry.speed<0&&entry.car.x<-60)entry.car.x=900;}}};
}
