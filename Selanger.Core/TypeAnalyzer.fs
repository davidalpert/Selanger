namespace Selanger.Core

open System
open System.Linq
open System.IO
open System.Reflection

type TypeSummary(file:FileInfo, t:Type) =
    member this.Name = t.Name
    member this.Namespace = t.Namespace
    member this.FilePath = file.FullName

type NamespaceSummary(asm:Assembly,ns:string,types:Type seq) = 
    member this.AssemblyName = System.IO.Path.GetFileName (asm.Location)
    member this.AssemblyVersion = asm.GetName().Version
    member this.Namespace = ns
    member this.Count = types.Count()

type TypeAnalyzer() = 
    let buildNamespaceSummary ((asm,ns),types:Type seq) =
        let t = types.FirstOrDefault()
        new NamespaceSummary(t.Assembly, ns, types)

    member this.Analyze([<ParamArray>] files:FileInfo[]) = 
        files |> Seq.where (fun(f) -> f.Exists) 
              |> Seq.map (fun(f) -> Assembly.LoadFile f.FullName)
              |> Seq.collect (fun(a) -> a.GetTypes()) 
              |> Seq.sortBy (fun (t:Type) -> t.Assembly.Location+"+"+t.Namespace+"+"+t.FullName)
              |> Seq.groupBy (fun(t:Type) -> t.Assembly)
              |> Seq.collect (fun(asm,types) -> types |> Seq.groupBy (fun(t) -> (asm,t.Namespace)))
              |> Seq.map buildNamespaceSummary
        //| 0 -> raise (new InvalidOperationException(String.Format("Could not find '{0}'.", file.FullName)))

