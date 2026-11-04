package mock

import (
 "context"
 "testing"

 "github.com/thanoslabs/forma/internal/adapters/agents"
)
func TestProvider(t *testing.T) {
 p:=New()
 list,err:=p.ListAgents(context.Background())
 if err!=nil||len(list)!=1||list[0].Status!=agents.Idle {t.Fatalf("initial agents: %v %v",list,err)}
 if !p.SetStatus("demo-agent",agents.Running) {t.Fatal("failed to update")}
 list,err=p.ListAgents(context.Background())
 if err!=nil||list[0].Status!=agents.Running {t.Fatalf("updated agents: %v %v",list,err)}
 if p.SetStatus("unknown",agents.Idle) {t.Fatal("updated missing agent")}
 ctx,cancel:=context.WithCancel(context.Background());cancel()
 if _,err:=p.ListAgents(ctx);err==nil {t.Fatal("expected context error")}
}
