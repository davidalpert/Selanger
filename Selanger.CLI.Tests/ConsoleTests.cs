using System;
using System.Diagnostics;
using System.IO;
using System.Reflection;
using ApprovalTests;
using NUnit.Framework;
using Selanger.CLI.Tests.Helpers;

namespace Selanger.CLI.Tests
{
    public class ConsoleTests
    {
        [Test]
        public void Get_help()
        {
            var args = new string[] {};

            var output = RunProgramAndCaptureOutput(args);

            Console.WriteLine(output);
        }

        [Test]
        public void SolutionReport()
        {
            var solutionFileDirectory = GetSolutionFileDirectory();

            var args = new[]
                {
                    solutionFileDirectory.FullName
                };

            var output = RunProgramAndCaptureOutput(args);

            Console.WriteLine(output);
        }

        [Test]
        public void ProjectGraphReport()
        {
            var solutionFileDirectory = GetSolutionFileDirectory();

            var args = new[]
                {
                    "-t", "g", solutionFileDirectory.FullName
                };

            var output = RunProgramAndCaptureOutput(args);

            Console.WriteLine(output);
        }

        [Test]
        [Ignore("TBD")]
        public void List_namespaces()
        {
            var source_dir = GetThisFileDirectory();
            var path_to_sln = Path.Combine(source_dir.FullName,"..","Selanger.sln");
            var sln_file = new FileInfo(path_to_sln);
            Assert.True(sln_file.Exists, sln_file.FullName);

            var directory = Path.GetDirectoryName(Assembly.GetExecutingAssembly().Location);
            var path_to_cli = Path.Combine(directory, "Selanger.exe");
            var cli = new FileInfo(path_to_cli);
            Assert.True(cli.Exists, cli.FullName);

            var args = new[]
                {
                    "-r",
                    "n",
                    "-i",
                    sln_file.FullName
                };

            string output;
            using (var cons = new RedirectedConsole())
            {
                Program.Main(args);

                output = cons.Output;
            }

            Console.WriteLine(output);
        }

        [Test]
        public void List_references()
        {
            var source_dir = GetThisFileDirectory();
            var path_to_sln = Path.Combine(source_dir.FullName,"..","Selanger.sln");
            var sln_file = new FileInfo(path_to_sln);
            Assert.True(sln_file.Exists, sln_file.FullName);

            var directory = Path.GetDirectoryName(Assembly.GetExecutingAssembly().Location);
            var path_to_cli = Path.Combine(directory, "Selanger.exe");
            var cli = new FileInfo(path_to_cli);
            Assert.True(cli.Exists, cli.FullName);

            var args = new[]
                {
                    "-r",
                    "r",
                    "-i",
                    sln_file.FullName,
                    new FileInfo(@"d:\obs\tfs\KORE Gen3 Services\Source\Gen3\ParsingService\ParsingService.sln").FullName
                };

            string output;
            using (var cons = new RedirectedConsole())
            {
                Program.Main(args);

                output = cons.Output;
            }

            Console.WriteLine(output);
            //Approvals.Verify(output);
        }

        private static string RunProgramAndCaptureOutput(string[] args)
        {
            string output;
            using (var cons = new RedirectedConsole())
            {
                Program.Main(args);

                output = cons.Output;
            }
            return output;
        }

        private DirectoryInfo GetThisFileDirectory()
        {
            var stacktrace = new StackTrace(true);
            var first_frame = stacktrace.GetFrame(0);
            var path_to_source = first_frame.GetFileName();
            var path_to_source_folder = Path.GetDirectoryName(path_to_source);
            return new DirectoryInfo(path_to_source_folder);
        }

        private static FileInfo GetSelengerExe()
        {
            var directory = Path.GetDirectoryName(Assembly.GetExecutingAssembly().Location);
            var path_to_cli = Path.Combine(directory, "Selanger.exe");
            var cli = new FileInfo(path_to_cli);
            Assert.True(cli.Exists, cli.FullName);
            return cli;
        }

        private DirectoryInfo GetSolutionFileDirectory()
        {
            var source_dir = GetThisFileDirectory();
            var path_to_sln = Path.Combine(source_dir.FullName, "..", "Selanger.sln");
            var sln_file = new FileInfo(path_to_sln);
            Assert.True(sln_file.Exists, sln_file.FullName);
            return sln_file.Directory;
        }
    }
}
