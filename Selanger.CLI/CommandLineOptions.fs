module CommandLineOptions

open System.IO

// pattern sourced from: http://fsharpforfunandprofit.com/posts/pattern-matching-command-line/

type ReportType =
| ProjectsReport
| StatisticsReport
| TreeReport

type CommandLineOptions = {
    directoryToScan: DirectoryInfo Option;
    outputFile: FileInfo Option;
    reportType: ReportType;
}

let print_help() =
    printfn ""
    printfn "Selanger [options] {folder_to_scan}"
    printfn ""
    printfn "Options:"
    printfn "-o|output {filepath} - path to the output file"
    printfn "-p|projects - run a project report"
    printfn "-s|stats    - run a statistics report"
    printfn "-t|tree     - run a tree report"
    printfn ""

// create the defaults
let defaultOptions = {
        directoryToScan = None;
        reportType = ProjectsReport;
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

    | "-p"::xs
    | "-projects"::xs ->
        parseCommandLineRec xs { optionsSoFar with reportType=ProjectsReport; }

    | "-s"::xs
    | "-stats"::xs ->
        parseCommandLineRec xs { optionsSoFar with reportType=StatisticsReport; }

    | "-t"::xs
    | "-tree"::xs ->
        parseCommandLineRec xs { optionsSoFar with reportType=TreeReport; }

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
