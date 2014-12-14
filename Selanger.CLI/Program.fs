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

    match opt.directoryToScan with
    | None -> print_help()
    | Some(dirToScan) ->

        printfn "Scanning: %s" dirToScan.FullName

        printfn ""

        let solutionFiles = dirToScan.GetFiles("*.sln", SearchOption.AllDirectories) |> List.ofArray


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

        let solutionList (slnFiles:FileInfo list) =
            slnFiles
            |> Seq.map (fun fi -> fi.FullName)
            |> Seq.iter (printfn "- %s")

        let projectGraph (fs:FileInfo list) =

            for f in fs do
                try
                    let sln = Solution.LoadFrom(f.FullName)
                    printfn "%s.sln" sln.Name
                    printfn "-------------------"
                    for p in sln.Projects do
                        let proj = p.Project
                        printfn "- %s [builds to: %s]" (Path.GetFileName(proj.FileName)) proj.AssemblyName
                        for a in p.Project.All<AssemblyReference>() do
                            let reference =
                                match a.Include.Contains(",") with
                                | true -> a.Include.Substring(0,a.Include.IndexOf(','))
                                | false -> a.Include
                            printfn "  --> %s" reference
                with
                | ex -> eprintfn "%s, Error loading solution, %s, %s" f.FullName ex.Message ex.StackTrace

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
        | SolutionReport-> solutionList solutionFiles
        | ProjectGraphReport -> projectGraph solutionFiles
        | NamespaceReport -> namespaceList solutionFiles
        | ReferencesReport -> referencesList solutionFiles

    0 // return an integer exit code
