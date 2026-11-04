package mock

import (
 "context"
 "github.com/thanoslabs/forma/internal/adapters/projects"
)

type Provider struct{}
func New() *Provider {return &Provider{}}
func (*Provider) ID() string{return "mock-projects"}
func (*Provider) Capabilities() []projects.Capability{return []projects.Capability{projects.Read}}
func (*Provider) ListProjects(ctx context.Context)([]projects.Project,error){
 if err:=ctx.Err();err!=nil{return nil,err}
 return []projects.Project{{ID:"demo",Name:"Demo Project"}},nil
}
func (*Provider) ListWorkItems(ctx context.Context,id string)([]projects.WorkItem,error){
 if err:=ctx.Err();err!=nil{return nil,err}
 if id!="demo"{return []projects.WorkItem{},nil}
 return []projects.WorkItem{{ID:"first-task",ProjectID:"demo",Title:"Welcome to Forma",Status:"open"}},nil
}
