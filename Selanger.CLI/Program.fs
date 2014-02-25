open System.IO
open System.Linq
open FubuCsProjFile
open Selanger.Core

open CommandLineOptions

[<EntryPoint>]
let public Main argv = 
    let opt = parseCommandLine argv

    let toFileInfo f = new FileInfo(f)

    let files = opt.files |> List.map toFileInfo

    match opt.files.Length with
    | 0 -> print_help
    | _ -> 

        let projectGraph (fs:FileInfo list) =
            printfn "Scanning:"
            for f in files do
               printfn "- %A" f.FullName

            printfn ""

            for f in fs do
                let sln = Solution.LoadFrom(f.FullName)
                printfn "%s" sln.Name
                printfn "-------------------"
                for p in sln.Projects do
                    let proj = p.Project
                    printfn "- %s -> %s @ %s" proj.ProjectName proj.AssemblyName proj.RootNamespace
                    for a in p.Project.All<AssemblyReference>() do
                        let reference =
                            match a.Include.Contains(",") with
                            | true -> a.Include.Substring(0,a.Include.IndexOf(','))
                            | false -> a.Include
                        printfn "  reference: %s" reference

        let namespaceList (fs:FileInfo list) =
            printfn "solution file, solution name, root namespace, project name, project file"
            for f in fs do
                let sln = Solution.LoadFrom(f.FullName)
                for sp in sln.Projects do
                    let proj = sp.Project
                    let project_path = (Path.Combine(sln.ParentDirectory, sp.RelativePath)) 
                    printfn "%s,%s,%s,%s,%s" sln.Filename sln.Name proj.RootNamespace proj.ProjectName project_path

        let referencesList (fs:FileInfo list) =
            printfn "solution file, solution name, project name, project file, assembly name, assembly reference"
            for f in fs do
                let sln = Solution.LoadFrom(f.FullName)
                for sp in sln.Projects do
                    let proj = sp.Project
                    let project_path = (Path.Combine(sln.ParentDirectory, sp.RelativePath)) 
                    for a in sp.Project.All<AssemblyReference>() do
                        let reference =
                            match a.Include.Contains(",") with
                            | true -> a.Include.Substring(0,a.Include.IndexOf(','))
                            | false -> a.Include
                        printfn "%s,%s,%s,%s,%s,%s" sln.Filename sln.Name proj.ProjectName project_path reference a.Include


        match opt.report with
        | ProjectGraphReport -> projectGraph files
        | NamespaceReport -> namespaceList files
        | ReferencesReport -> referencesList files

    0 // return an integer exit code
