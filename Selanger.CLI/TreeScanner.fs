﻿module TreeScanner

open System
open System.IO
open FubuCsProjFile
open Helpers

type ProjectNamingConvention =
| MatchesFolderName
| DoesNotMatchFolderName

type AssemblyNamingConvention =
| MatchesProjectName
| DoesNotMatchProjectName

let Scan (dirToScan:DirectoryInfo) (out:FileInfo option) =

    let printAssemblyReference (r:AssemblyReference) =
        let specificVersion = match r.SpecificVersion.HasValue && r.SpecificVersion.Value with
                              | true -> "specific: "
                              | false -> ""
        let msg = match String.IsNullOrWhiteSpace(r.HintPath) with
                 | true  -> sprintf "       - %s" r.Include
                 | false -> sprintf "       - %s (%s%s)" r.Include specificVersion r.HintPath
        writen msg out

    let printProjectReference (r:ProjectReference) =
        let msg = sprintf "       + %s" r.ProjectName
        writen msg out

    let printDependencies (p:SolutionProject) =

        p.Project.All<AssemblyReference>()
        |> Seq.sortBy (fun r -> r.Include.ToLowerInvariant())
        |> Seq.iter printAssemblyReference

        p.Project.All<ProjectReference>()
        |> Seq.sortBy (fun r -> r.ProjectName.ToLowerInvariant())
        |> Seq.iter printProjectReference

    let printProjectNode (p:SolutionProject) =
        let projectName = p.ProjectName
        let pathToProject = Path.GetDirectoryName(p.RelativePath.Replace('/','\\'))
        let language = match Path.GetExtension(p.RelativePath).ToLowerInvariant() with
                       | ".csproj" -> "C#"
                       | ".fsproj" -> "F#"
                       | ".vbproj" -> "VB"
                       | s         -> s

        let projectNamingConvention  = match projectName with
                                       | pathToProject -> MatchesFolderName
                                       | _             -> DoesNotMatchFolderName

        let assemblyNamingConvention = match p.Project.AssemblyName with
                                       | projectName   -> MatchesProjectName
                                       | _             -> DoesNotMatchProjectName

        let msg = match (projectNamingConvention, assemblyNamingConvention) with
                      | (MatchesFolderName,       MatchesProjectName)      -> sprintf "  [%s] %s" language p.ProjectName
                      | (DoesNotMatchFolderName,  MatchesProjectName)      -> sprintf "  [%s] %s (%s)" language p.ProjectName pathToProject
                      | (MatchesFolderName,       DoesNotMatchProjectName) -> sprintf "  [%s] %s (=> %s)" language p.ProjectName p.Project.AssemblyName
                      | (DoesNotMatchFolderName,  DoesNotMatchProjectName) -> sprintf "  [%s] %s (%s => %s)" language p.ProjectName (Path.GetDirectoryName(p.RelativePath.Replace('/','\\'))) p.Project.AssemblyName
        writen msg out

        printDependencies p

    let solutionFolderId = new Guid("2150E333-8FDC-42A3-9474-1A3956D46DE8")

    let printSolutionNode (f:FileInfo) =
        let pathToSolution = relativePath dirToScan (Some(f))
        match pathToSolution with
        | "" -> writen (sprintf "- %s" (fileName (Some(f)))) out
        | _  -> writen (sprintf "- %s [%s]" (fileName (Some(f))) (relativePath dirToScan (Some(f)))) out

        try
            let sln = Solution.LoadFrom(f.FullName)
            sln.Projects
            |> Seq.filter (fun p -> p.ProjectType <> solutionFolderId)
            |> Seq.iter printProjectNode
        with
        | ex ->
            let msg = sprintf "%s%s%s" ex.Message Environment.NewLine ex.StackTrace
            match out with
            | Some(f) -> eprintfn "%s" msg
            | None    -> writen msg out

    let solutionFiles = dirToScan.GetFiles("*.sln", SearchOption.AllDirectories) |> List.ofArray

    writen (sprintf "%s" dirToScan.FullName) out
    writen "" out

    solutionFiles
    |> Seq.iter printSolutionNode
