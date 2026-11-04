import { useEffect, useRef } from "react";
import { Application, Container, Graphics, Text } from "pixi.js";

export function Office() {
  const host = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const element = host.current;
    if (!element) return;
    const app = new Application();
    let disposed = false;
    let actor: Container | undefined;
    let elapsed = 0;
    const init = async () => {
      await app.init({ width: 860, height: 490, background: "#c4e5ed", antialias: true, resolution: Math.min(window.devicePixelRatio || 1, 2), autoDensity: true });
      if (disposed) { app.destroy(true); return; }
      element.appendChild(app.canvas);
      const room = new Graphics().roundRect(40, 36, 780, 418, 28).fill("#f5dfb6").stroke({ color: "#64818a", width: 7 });
      app.stage.addChild(room);
      const carpet = new Graphics().ellipse(430, 310, 290, 115).fill("#a9c8a6");
      app.stage.addChild(carpet);
      for (const x of [220, 430, 640]) {
        const desk = new Graphics().roundRect(x - 65, 205, 130, 50, 12).fill("#9b684e").roundRect(x - 35, 170, 70, 48, 6).fill("#3d5664");
        app.stage.addChild(desk);
      }
      actor = new Container();
      const body = new Graphics().ellipse(0, 0, 29, 36).fill("#7e70cf").ellipse(-9, -8, 4, 6).fill("#ffffff").ellipse(9, -8, 4, 6).fill("#ffffff").ellipse(-9, -8, 2, 3).fill("#25384d").ellipse(9, -8, 2, 3).fill("#25384d").ellipse(0, 5, 13, 7).fill("#e5ba67");
      actor.addChild(body);
      actor.position.set(430, 355);
      actor.eventMode = "static";
      actor.cursor = "pointer";
      actor.on("pointertap", () => element.dispatchEvent(new CustomEvent("forma-agent-select", { bubbles: true, detail: { agentId: "demo-agent" } })));
      app.stage.addChild(actor);
      const label = new Text({ text: "Demo Agent · Idle", style: { fontFamily: "sans-serif", fontSize: 15, fill: "#28384a" } });
      label.anchor.set(0.5);
      label.position.set(430, 410);
      app.stage.addChild(label);
      app.ticker.add(ticker => { elapsed += ticker.deltaTime / 60; if (actor) actor.y = 355 + Math.sin(elapsed * 2) * 3; });
    };
    void init();
    return () => { disposed = true; if (app.renderer) app.destroy(true); };
  }, []);
  return <div ref={host} className="office-canvas" aria-label="Illustrated office with three desks and one selectable demo agent" role="img" />;
}
