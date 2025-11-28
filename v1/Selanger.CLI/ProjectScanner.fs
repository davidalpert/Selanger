module ProjectScanner

open System
open System.IO
open FubuCsProjFile
open Helpers

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

let Scan (dirToScan:DirectoryInfo) (out:FileInfo option) =

    let solutionFiles = dirToScan.GetFiles("*.sln", SearchOption.AllDirectories) |> List.ofArray

    let solutionFolderId = new Guid("2150E333-8FDC-42A3-9474-1A3956D46DE8")

    writen ScanRecord.HeadingRow out

    for f in solutionFiles do

        let slnRecord = {
            RootPath = dirToScan;
            SolutionPath = Some(f);
            RelativeProjectPath = "";
            ProjectFile = None;
            Error = None;
        }

        try
            let sln = Solution.LoadFrom(f.FullName)

            let projectsToScan = sln.Projects
                                 |> Seq.filter (fun p -> p.ProjectType <> solutionFolderId)

            for p in projectsToScan do
                let proj = p.Project
                let projRecord = { slnRecord with
                                        RelativeProjectPath = p.RelativePath
                                        ProjectFile = Some(p.Project);
                                 }
                writen projRecord.serialized out
        with
        | ex ->
            let errorRecord = { slnRecord with Error = Some(ex);}
            writen errorRecord.serialized out
            eprintfn "%s" errorRecord.serialized
