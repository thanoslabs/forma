import { describe,expect,it } from "vitest";
import { defaultLayout,validateLayout } from "./layout";
describe("layout validation",()=>{
 it("rejects invalid or unsupported data",()=>{
  expect(validateLayout(null)).toEqual(defaultLayout);
  expect(validateLayout({version:2,desks:[]})).toEqual(defaultLayout);
  expect(validateLayout({version:1,desks:[{x:NaN,y:200}]})).toEqual(defaultLayout);
 });
 it("accepts valid desk coordinates",()=>{
  expect(validateLayout(defaultLayout)).toEqual(defaultLayout);
 });
});
