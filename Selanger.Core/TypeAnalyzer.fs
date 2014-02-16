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
    let buildNamespaceSummary asm (ns,types) =
        new NamespaceSummary(asm, ns, types)

    member this.Analyze(file:FileInfo) = 
        if file.Exists then 
            let asm = Assembly.LoadFile file.FullName
            asm.GetTypes()
                |> Seq.ofArray
                |> Seq.sortBy (fun (t:Type) -> t.Namespace+"+"+t.FullName)
                |> Seq.groupBy (fun (t:Type) -> t.Namespace) 
                |> Seq.map (buildNamespaceSummary asm)
        else
            raise (new InvalidOperationException(String.Format("Could not find '{0}'.", file.FullName)))

