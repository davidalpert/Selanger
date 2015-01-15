module NugetPackageScanner

open System
open System.IO
open System.Text.RegularExpressions
open FubuCsProjFile
open Helpers

type ProjectNamingConvention =
| MatchesFolderName
| DoesNotMatchFolderName

type AssemblyNamingConvention =
| MatchesProjectName
| DoesNotMatchProjectName

type NugetPackageDependency(name:string, version:Version) =
    member this.Name = name
    member this.Version = version

type HintPathRegexKey() =
    static member Match = 0
    static member PackageName = 1
    static member VersionText = 5

let Scan (dirToScan:DirectoryInfo) (out:FileInfo option) =

    let _reHintPath = new Regex(@"\\packages\\(([a-zA-Z][a-zA-Z0-9]*)(.([a-zA-Z][a-zA-Z0-9]*))*).(\d+(.\d+){0,3})\\")

    let parseHintPath(hintPath:string) =
        match (String.IsNullOrWhiteSpace(hintPath)) with
        | true -> None
        | false ->
            let result = _reHintPath.Match(hintPath)
            match result.Success with
            | false -> None
            | true  ->
                let packageName = result.Groups.[HintPathRegexKey.PackageName].Value
                let versionText = result.Groups.[HintPathRegexKey.VersionText].Value
                let versionNumber = new Version(versionText)
                Some(NugetPackageDependency(packageName,versionNumber))

    let printAssemblyReference (r:AssemblyReference) =
        let specificVersion = match r.SpecificVersion.HasValue && r.SpecificVersion.Value with
                              | true -> "specific: "
                              | false -> ""
        let msg = match String.IsNullOrWhiteSpace(r.HintPath) with
                 | true  -> sprintf "       - %s" r.Include
                 | false -> sprintf "       - %s (%s%s)" r.Include specificVersion r.HintPath
        parseHintPath(r.HintPath)

    let printDependencies (p:SolutionProject) =
        try
            p.Project.All<AssemblyReference>()
            |> Seq.sortBy (fun r -> r.Include.ToLowerInvariant())
            |> Seq.map printAssemblyReference
            |> Seq.where (fun o -> o.IsSome)
            |> Seq.map (fun o -> o.Value)
        with
        | ex ->
            let msg = sprintf "%s%s%s" ex.Message Environment.NewLine ex.StackTrace
            match out with
            | Some(f) -> eprintfn "%s" msg
            | None    -> writen msg out
            Seq.empty

    let solutionFolderId = new Guid("2150E333-8FDC-42A3-9474-1A3956D46DE8")

    let getPackageDependenciesFromSolution (f:FileInfo) =
        (*
        let pathToSolution = relativePath dirToScan (Some(f))
        match pathToSolution with
        | "" -> writen (sprintf "- %s" (fileName (Some(f)))) out
        | _  -> writen (sprintf "- %s [%s]" (fileName (Some(f))) (relativePath dirToScan (Some(f)))) out
        *)
        try
            let sln = Solution.LoadFrom(f.FullName)
            sln.Projects
            |> Seq.filter (fun p -> p.ProjectType <> solutionFolderId)
            |> Seq.collect printDependencies
        with
        | ex ->
            let msg = sprintf "%s%s%s" ex.Message Environment.NewLine ex.StackTrace
            match out with
            | Some(f) -> eprintfn "%s" msg
            | None    -> writen msg out
            Seq.empty

    let solutionFiles = dirToScan.GetFiles("*.sln", SearchOption.AllDirectories) |> List.ofArray

    writen (sprintf "%s" dirToScan.FullName) out
    writen "" out

    let dependenciesByName =
        solutionFiles
        |> Seq.collect getPackageDependenciesFromSolution
        |> Seq.groupBy (fun d -> d.Name)


    dependenciesByName
    |> Seq.collect (fun (name,versions) -> versions |> Seq.distinctBy (fun d -> d.Version))
    |> Seq.sortBy (fun d -> d.Name)
    |> Seq.iter (fun d -> writen (sprintf "%s [%s]" d.Name (d.Version.ToString())) out)

    writen "" out

    dependenciesByName
    |> Seq.map (fun (name,versions) -> name)
    |> Seq.sortBy (fun name -> name)
    |> Seq.iter (fun name -> writen name out)

