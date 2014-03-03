open System
open System.IO
open System.Linq
open FubuCsProjFile
open Selanger.Core

open Microsoft.FSharp.Core.Printf

open CommandLineOptions

[<EntryPoint>]
let public Main argv = 
    let opt = parseCommandLine argv

    let toFileInfo f = new FileInfo(f)

    let files = opt.files |> List.map toFileInfo

    match opt.files.Length with
    | 0 -> print_help()
    | _ -> 

        printfn "Scanning:"
        for f in files do
           printfn "- %A" f.FullName

        printfn ""

        let appendn (file:FileInfo) (line:string) = 
          use wr = StreamWriter(file.FullName, true)
          wr.WriteLine(line)

        if opt.outputFile.IsSome then
            // ensure that the file is empty
            opt.outputFile.Value.Delete()

        let writen (s:string) = 
            match opt.outputFile with
                | Some(file) -> appendn file s
                | None -> printf "%s" s

        let writefn fmt = Printf.kprintf writen fmt

        let projectGraph (fs:FileInfo list) =

            for f in fs do
                try 
                    let sln = Solution.LoadFrom(f.FullName)
                    writefn "%s" sln.Name
                    writen "-------------------"
                    for p in sln.Projects do
                        let proj = p.Project
                        writefn "- %s -> %s @ %s" proj.ProjectName proj.AssemblyName proj.RootNamespace
                        for a in p.Project.All<AssemblyReference>() do
                            let reference =
                                match a.Include.Contains(",") with
                                | true -> a.Include.Substring(0,a.Include.IndexOf(','))
                                | false -> a.Include
                            writefn "  reference: %s" reference
                with
                | ex -> writefn "%s, Error loading solution, %s, %s" f.FullName ex.Message ex.StackTrace

        let namespaceList (fs:FileInfo list) =
            writen "solution file, solution name, root namespace, project name, project file"
            for f in fs do
                try 
                    let sln = Solution.LoadFrom(f.FullName)
                    for sp in sln.Projects do
                        let proj = sp.Project
                        let project_path = (Path.Combine(sln.ParentDirectory, sp.RelativePath)) 
                        writefn "%s,%s,%s,%s,%s" sln.Filename sln.Name proj.RootNamespace proj.ProjectName project_path
                with
                | ex -> writefn "%s, Error loading solution, %s, %s" f.FullName ex.Message (ex.StackTrace.Replace(Environment.NewLine, "--NEWLINE--"))

        let referencesList (fs:FileInfo list) =
            writen "solution file, solution name, project name, project file, assembly name, assembly reference"
            for f in fs do
                try 
                    let sln = Solution.LoadFrom(f.FullName)
                    for sp in sln.Projects do
                        let proj = sp.Project
                        let project_path = (Path.Combine(sln.ParentDirectory, sp.RelativePath)) 
                        for a in sp.Project.All<AssemblyReference>() do
                            let reference =
                                match a.Include.Contains(",") with
                                | true -> a.Include.Substring(0,a.Include.IndexOf(','))
                                | false -> a.Include
                            writefn "%s,%s,%s,%s,%s,%s" sln.Filename sln.Name proj.ProjectName project_path reference a.Include
                with
                | ex -> writefn "%s, Error loading solution, %s, %s" f.FullName ex.Message (ex.StackTrace.Replace(Environment.NewLine, "--NEWLINE--"))


        match opt.report with
        | ProjectGraphReport -> projectGraph files
        | NamespaceReport -> namespaceList files
        | ReferencesReport -> referencesList files

    0 // return an integer exit code
