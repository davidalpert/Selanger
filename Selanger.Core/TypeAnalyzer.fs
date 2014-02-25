namespace Selanger.Core

open System
open System.Linq
open System.IO
open System.Reflection

type TypeSummary(t:Type) =
    member this.Name = t.Name
    member this.Namespace = t.Namespace
    member this.FileName = Path.GetFileName t.Assembly.Location
    member this.FileDirectory = Path.GetDirectoryName t.Assembly.Location
    member this.FileVersion = t.Assembly.GetName().Version

type NamespaceSummary(asm:Assembly,ns:string,types:Type seq) = 
    member this.AssemblyName = System.IO.Path.GetFileName (asm.Location)
    member this.AssemblyVersion = asm.GetName().Version
    member this.Namespace = ns
    member this.Count = types.Count()

type TypeAnalyzer() = 
    let buildNamespaceSummary ((asm,ns),types:Type seq) =
        let t = types.FirstOrDefault()
        new NamespaceSummary(t.Assembly, ns, types)

    let scanForTypes([<ParamArray>] files:FileInfo[]) = 
        let a = files |> Seq.where (fun f -> f.Exists) 
        let b = a      |> Seq.map (fun f -> 
                            try 
                                Some (Assembly.LoadFile f.FullName)
                            finally
                                None |> ignore
                         )
        let c = b      |> Seq.where (fun (a:Assembly option) -> a.IsSome)
        let d = c      |> Seq.map (fun (a:Assembly option) -> a.Value)
        let e = d      |> Seq.map (fun(a:Assembly) -> 
                            try 
                                Some (a.GetTypes())
                            finally
                                None |> ignore
                         ) 
        let f = e      |> Seq.where (fun (a:Type[] option) -> a.IsSome)
        let g = f      |> Seq.collect (fun (a:Type[] option) -> a.Value)
        let h = g      |> Seq.sortBy (fun (t:Type) -> t.Assembly.Location+"+"+t.Namespace+"+"+t.FullName)
        h
        
    member this.Analyze([<ParamArray>] files:FileInfo[]) = 
        scanForTypes(files)
              |> Seq.map (fun t -> new TypeSummary(t))


    member this.SummarizeNamespaces([<ParamArray>] files:FileInfo[]) = 
        scanForTypes(files)
              |> Seq.groupBy (fun(t:Type) -> t.Assembly)
              |> Seq.collect (fun(asm,types) -> types |> Seq.groupBy (fun(t) -> (asm,t.Namespace)))
              |> Seq.map buildNamespaceSummary
        //| 0 -> raise (new InvalidOperationException(String.Format("Could not find '{0}'.", file.FullName)))

