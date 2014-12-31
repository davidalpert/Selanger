open System
open System.IO
open System.Linq
open FubuCsProjFile
open Selanger.Core

open Microsoft.FSharp.Core.Printf

open CommandLineOptions

let relativePath (root:DirectoryInfo) (rel:FileInfo option) =
    match rel with
    | Some(fi) -> fi.Directory.FullName.Substring(root.FullName.Length)
    | None -> "n/a"

let fileName (f:FileInfo option) =
    match f with
    | Some(fi) -> Path.GetFileName(fi.FullName)
    | None -> "n/a"

type ScanRecord =
    {
        RootPath : DirectoryInfo;
        SolutionPath : FileInfo option;
        RelativeProjectPath : string;
        ProjectFile : CsProjFile option;
        Error : Exception option;
    }

    static member Headings = [
            "Root Path"
            "Relative Path to Solution"
            "Solution Name"
            "Relative Path to Project"
            "Project Name"
            "AssemblyName"
        ]

    static member HeadingRow = String.Join(",", ScanRecord.Headings)

    member this.serialized = match this.Error with
                             | Some(ex) -> String.Join(",", [
                                                                this.RootPath.FullName
                                                                relativePath this.RootPath this.SolutionPath
                                                                fileName this.SolutionPath
                                                                ex.Message
                                                                ex.StackTrace
                                                            ])
                             | None ->     let projectFile = this.ProjectFile.Value
                                           String.Join(",", [
                                                                this.RootPath.FullName
                                                                relativePath this.RootPath this.SolutionPath
                                                                fileName this.SolutionPath
                                                                Path.GetDirectoryName(this.RelativeProjectPath)
                                                                Path.GetFileName(projectFile.FileName)
                                                                projectFile.AssemblyName
                                                            ])

[<EntryPoint>]
let public Main argv =
    let opt = parseCommandLine argv

    match opt.directoryToScan with
    | None -> print_help()
    | Some(dirToScan) ->

        if opt.outputFile.IsSome then
            // ensure that the file is empty
            opt.outputFile.Value.Delete()

        let solutionFiles = dirToScan.GetFiles("*.sln", SearchOption.AllDirectories) |> List.ofArray

        // appends a line to a file
        let appendn (file:FileInfo) (line:string) =
          use wr = new StreamWriter(file.FullName, true)
          wr.WriteLine(line)

        // writes a line to the chosen output stream (i.e. file or console)
        let writen (s:string) =
            match opt.outputFile with
                | Some(file) -> appendn file s
                | None -> printfn "%s" s


        writen ScanRecord.HeadingRow

        for f in solutionFiles do

            let slnRecord = {
                RootPath = dirToScan;
                SolutionPath = Some(f);
                RelativeProjectPath = "";
                ProjectFile = None
                Error = None;
            }

            try
                let sln = Solution.LoadFrom(f.FullName)
                for p in sln.Projects do
                    let proj = p.Project
                    let projRecord = { slnRecord with
                                            RelativeProjectPath = p.RelativePath;
                                            ProjectFile = Some(p.Project);
                                     }
                    writen projRecord.serialized
            with
            | ex ->
                let errorRecord = { slnRecord with Error = Some(ex);}
                eprintfn "%s" errorRecord.serialized

    0 // return an integer exit code

        (*

        let writefn fmt = Printf.kprintf writen fmt

        let solutionList (slnFiles:FileInfo list) =
            slnFiles
            |> Seq.map (fun fi -> fi.FullName)
            |> Seq.iter (printfn "- %s")
        *)


        (*
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
        *)