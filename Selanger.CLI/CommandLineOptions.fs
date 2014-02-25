module CommandLineOptions

// pattern sourced from: http://fsharpforfunandprofit.com/posts/pattern-matching-command-line/

type OutputFormatOption = OutputCSV //| OutputXML
type ReportTypeOption = ProjectGraphReport | NamespaceReport | ReferencesReport

type CommandLineOptions = {
    files: string list;
    outputFormat : OutputFormatOption;
    report: ReportTypeOption;
}

let print_help =
    printfn ""
    printfn "Selanger [options] -i solution1.sln solution2.sln ..."
    printfn ""
    printfn "Options:"
    printfn "-r|report {reportType} - configures the report type"
    printfn ""
    printfn "Report Types (default: projectGraph)"
    printfn " n|namespaces - list the namespaces in the included projects"
    printfn " r|references - list the assembly referenced by the included projects"
    printfn " pg|projectGraph - dump a graph of the solution/project structure"
    printfn ""

// create the defaults
let defaultOptions = {
        files = [];
        outputFormat = OutputCSV;
        report = ProjectGraphReport;
    }

// create the "helper" recursive function
let rec parseCommandLineRec args optionsSoFar = 
    match args with 

    | "-i"::filePaths 
    | "-input"::filePaths ->
        { optionsSoFar with files=filePaths }

    | "-r"::xs 
    | "-report"::xs ->
        //start a submatch on the next arg
        match xs with
        | "pg"::xss 
        | "projectGraph"::xss -> 
            parseCommandLineRec xss { optionsSoFar with report=ProjectGraphReport}

        | "n"::xss 
        | "namespaces"::xss -> 
            parseCommandLineRec xss { optionsSoFar with report=NamespaceReport}

        | "r"::xss 
        | "references"::xss -> 
            parseCommandLineRec xss { optionsSoFar with report=ReferencesReport}

        // handle unrecognized option and keep looping
        | _ -> 
            eprintfn "ReportType needs a second argument (n|namespaces)"
            parseCommandLineRec xs optionsSoFar 

    // handle unrecognized option and keep looping
    | x::xs -> 
        eprintfn "Option '%s' is unrecognized" x
        parseCommandLineRec xs optionsSoFar 

    // empty list means we're done.
    | [] -> 
       optionsSoFar  

// create the "public" parse function
let parseCommandLine (args:string[]) = 
    // call the recursive one with the initial options
    let argList = List.ofSeq args
    parseCommandLineRec argList defaultOptions 
