module StatisticsScanner

open System
open System.IO
open FubuCsProjFile
open Helpers

let Scan (dirToScan:DirectoryInfo) (out:FileInfo option) =

    let solutionFiles = dirToScan.GetFiles("*.sln", SearchOption.AllDirectories) |> List.ofArray

    let solutionFolderId = new Guid("2150E333-8FDC-42A3-9474-1A3956D46DE8")

    writen (sprintf "Solutions: %i" solutionFiles.Length) out

    let projectsToScan =
        solutionFiles
        |> List.collect (fun f ->
                             try
                                let sln = Solution.LoadFrom(f.FullName)
                                sln.Projects
                                 |> Seq.filter (fun p -> p.ProjectType <> solutionFolderId)
                                 |> List.ofSeq
                             with
                             | ex ->
                                let msg = sprintf "%s%s%s" ex.Message Environment.NewLine ex.StackTrace
                                match out with
                                | Some(f) -> eprintfn "%s" msg
                                | None    -> writen msg out
                                []
                        )

    writen (sprintf "Projects: %i" projectsToScan.Length) out
