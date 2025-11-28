module Helpers

open System.IO

// appends a line to a file
let appendn (file:FileInfo) (line:string) =
  use wr = new StreamWriter(file.FullName, true)
  wr.WriteLine(line)

// writes a line to the chosen output stream (i.e. file or console)
let writen (s:string) (f:FileInfo option) =
    match f with
        | Some(file) -> appendn file s
        | None -> printfn "%s" s

let relativePath (root:DirectoryInfo) (rel:FileInfo option) =
    match rel with
    | Some(fi) -> fi.Directory.FullName.Substring(root.FullName.Length)
    | None -> "n/a"

let fileName (f:FileInfo option) =
    match f with
    | Some(fi) -> Path.GetFileName(fi.FullName)
    | None -> "n/a"

