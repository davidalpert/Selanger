open System.IO
open Selanger.Core

open CommandLineOptions

[<EntryPoint>]
let main argv = 
    let opt = parseCommandLine argv

    printfn "Scanning:"
    for f in opt.files do
       printfn "- %A" f

    printfn ""

    let toFileInfo f = new FileInfo(f)

    let getAllAssembliesUnderDirectory(f) = 
        let dlls = Directory.GetFiles(f, "*.dll", SearchOption.AllDirectories)
        let exes = Directory.GetFiles(f, "*.exe", SearchOption.AllDirectories)
        Array.concat [|dlls;exes|]
        |> Array.map toFileInfo
        |> List.ofArray

    let filelist = opt.files 
                 |> List.collect (fun f -> 
                               match File.Exists(f) with
                               | true -> [new FileInfo(f)]
                               | false -> match Directory.Exists(f) with
                                          | true -> getAllAssembliesUnderDirectory(f)
                                          | false -> []
                              )
                 |> Array.ofList

    let analyzer = new TypeAnalyzer()

    match opt.outputContent with
    | OutputTypeSummary ->
        let result = analyzer.Analyze(filelist)

        printfn "namespace,type name,assembly name,assembly version,location"
        for t in result do
            printfn "%A,%A,%A,%A,%A" t.Namespace t.Name t.FileName t.FileVersion t.FileDirectory

    | OutputNamespaceSummary ->
        let result = analyzer.SummarizeNamespaces(filelist)

        for ns in result do
            printfn "- %A (%A) [%A, %A]" ns.Namespace ns.Count ns.AssemblyName ns.AssemblyVersion
        
    0 // return an integer exit code
