module CommandLineOptions

open System.IO

// pattern sourced from: http://fsharpforfunandprofit.com/posts/pattern-matching-command-line/

type OutputFormatOption = OutputCSV //| OutputXML
type ReportTypeOption =
| SolutionReport
| ProjectGraphReport
| NamespaceReport
| ReferencesReport

type CommandLineOptions = {
    directoryToScan: DirectoryInfo Option;
    outputFormat : OutputFormatOption;
    report: ReportTypeOption;
    outputFile: FileInfo Option;
}

let print_help() =
    printfn ""
    printfn "Selanger [options] {folder_to_scan}"
    printfn ""
    printfn "Options:"
    printfn "-o|output {filepath} - path to the output file"
    printfn "-t|type {report type} - type of report to issue"
    printfn ""
    printfn "Report Types (default: projectGraph)"
    printfn " s|solutions - list the solutions under the scan"
    //printfn " n|namespaces - list the namespaces in the included projects"
    //printfn " r|references - list the assembly referenced by the included projects"
    printfn " pg|projectGraph - dump a graph of the solution/project structure"
    printfn ""

// create the defaults
let defaultOptions = {
        directoryToScan = None;
        outputFormat = OutputCSV;
        report = SolutionReport;
        outputFile = None;
    }

// create the "helper" recursive function
let rec parseCommandLineRec args optionsSoFar =
    match args with

    | "-o"::xs
    | "-output"::xs ->
        //start a submatch on the next arg
        match xs with
        | file_path::xss ->
            parseCommandLineRec xss { optionsSoFar with outputFile=Some(new FileInfo(file_path)) }

    | "-t"::xs
    | "-type"::xs ->
        //start a submatch on the next arg
        match xs with

        | "s"::xss
        | "solutions"::xss ->
            parseCommandLineRec xss { optionsSoFar with report=SolutionReport}

        | "g"::xss
        | "projectGraph"::xss ->
            parseCommandLineRec xss { optionsSoFar with report=ProjectGraphReport}

        // handle unrecognized option and keep looping
        | _ ->
            eprintfn "ReportType needs a second argument."
            parseCommandLineRec xs optionsSoFar

    // handle unrecognized option and keep looping
    | x::xs ->
        match Directory.Exists(x) with
        | true -> parseCommandLineRec xs { optionsSoFar with directoryToScan = Some(new DirectoryInfo(x)) }
        | false -> eprintfn "'%s' is not a valid directory name" x
                   parseCommandLineRec xs optionsSoFar

    // empty list means we're done.
    | [] ->
       optionsSoFar

// the "public" parse function
let parseCommandLine (args:string[]) =
    // call the recursive one with the initial options
    let argList = List.ofSeq args
    parseCommandLineRec argList defaultOptions
