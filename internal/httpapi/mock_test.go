package httpapi

import (
 "bytes"
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "testing"

 "github.com/thanoslabs/forma/internal/adapters/agents"
)
func TestMockAPI(t *testing.T){
 mux:=http.NewServeMux();NewMockAPI().Register(mux)
 for _,path:=range []string{"/api/v1/demo/agents","/api/v1/demo/projects","/api/v1/demo/projects/demo/items"}{
  r:=httptest.NewRecorder();mux.ServeHTTP(r,httptest.NewRequest("GET",path,nil))
  if r.Code!=200{t.Fatalf("%s: %d",path,r.Code)}
 }
 body,_:=json.Marshal(map[string]string{"status":"running"})
 req:=httptest.NewRequest("POST","/api/v1/demo/agents/demo-agent/status",bytes.NewReader(body))
 req.Header.Set("Content-Type","application/json")
 r:=httptest.NewRecorder();mux.ServeHTTP(r,req)
 if r.Code!=200{t.Fatalf("status update: %d",r.Code)}
 r=httptest.NewRecorder();mux.ServeHTTP(r,httptest.NewRequest("GET","/api/v1/demo/agents",nil))
 var list []agents.Agent
 if err:=json.Unmarshal(r.Body.Bytes(),&list);err!=nil||len(list)!=1||list[0].Status!=agents.Running{t.Fatalf("agents: %v %v",list,err)}
}
func TestInvalidStatus(t *testing.T){
 mux:=http.NewServeMux();NewMockAPI().Register(mux)
 req:=httptest.NewRequest("POST","/api/v1/demo/agents/demo-agent/status",bytes.NewBufferString(`{"status":"bogus"}`))
 req.Header.Set("Content-Type","application/json")
 r:=httptest.NewRecorder();mux.ServeHTTP(r,req)
 if r.Code!=400{t.Fatalf("expected 400, got %d",r.Code)}
}
