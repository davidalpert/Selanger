module CommandLineOptions

// pattern sourced from: http://fsharpforfunandprofit.com/posts/pattern-matching-command-line/

type OrderByOption = OrderByNamespace | OrderByAssembly
type SubdirectoryOption = IncludeSubdirectories | ExcludeSubdirectories
type OutputFormatOption = OutputCSV
type OutputContentOption = OutputNamespaceSummary | OutputTypeSummary

type CommandLineOptions = {
    subdirectories: SubdirectoryOption;
    orderBy: OrderByOption;
    files: string list;
    outputContent : OutputContentOption;
    outputFormat : OutputFormatOption;
}

// create the defaults
let defaultOptions = {
        subdirectories = ExcludeSubdirectories;
        orderBy = OrderByNamespace;
        files = [];
        outputContent = OutputTypeSummary;
        outputFormat = OutputCSV;
    }

// create the "helper" recursive function
let rec parseCommandLineRec args optionsSoFar = 
    match args with 

    | "-f"::xs 
    | "-format"::xs ->
        //start a submatch on the next arg
        match xs with
        | "c"::xss 
        | "csv"::xss -> 
            parseCommandLineRec xss { optionsSoFar with outputFormat=OutputCSV}

        // handle unrecognized option and keep looping
        | _ -> 
            eprintfn "OutputFormat needs a second argument (c|csv)"
            parseCommandLineRec xs optionsSoFar 

    | "-t"::xs 
    | "-type"::xs ->
        //start a submatch on the next arg
        match xs with
        | "t"::xss 
        | "types"::xss -> 
            parseCommandLineRec xss { optionsSoFar with outputContent=OutputTypeSummary}

        | "n"::xss 
        | "namespaces"::xss -> 
            parseCommandLineRec xss { optionsSoFar with outputContent=OutputNamespaceSummary}

        // handle unrecognized option and keep looping
        | _ -> 
            eprintfn "OutputContent needs a second argument (c|csv)"
            parseCommandLineRec xs optionsSoFar 

    | "-i"::filePaths 
    | "-input"::filePaths ->
        { optionsSoFar with files=filePaths }

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

    // parsing for other pending options
    (*
    // match verbose flag
    | "/v"::xs -> 
        let newOptionsSoFar = { optionsSoFar with verbose=VerboseOutput}
        parseCommandLineRec xs newOptionsSoFar 
    *)

    (*
    // match subdirectories flag
    | "-r"::xs 
    | "-recursive"::xs -> 
        parseCommandLineRec xs { optionsSoFar with subdirectories=IncludeSubdirectories}

    // match sort order flag
    | "-s"::xs 
    | "-sort"::xs -> 
        //start a submatch on the next arg
        match xs with
        | "n"::xss 
        | "namespace"::xss -> 
            parseCommandLineRec xss { optionsSoFar with orderBy=OrderByNamespace}
        | "a"::xss 
        | "assembly"::xss -> 
            parseCommandLineRec xss { optionsSoFar with orderBy=OrderByAssembly}
        // handle unrecognized option and keep looping
        | _ -> 
            printfn "SortBy needs a second argument (n|namespace or a|assembly)"
            parseCommandLineRec xs optionsSoFar 
    *)