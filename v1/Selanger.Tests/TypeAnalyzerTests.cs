using System;
using System.IO;
using System.Linq;
using ApprovalTests;
using NUnit.Framework;
using Selanger.Core;

namespace Selanger.Tests
{
    [TestFixture]
    public class TypeAnalyzerTests
    {
        [Test]
        public void CanGenerateTypeReport()
        {
            var asm = @".\nunit.framework.dll";
            var file = new FileInfo(asm);
            var reporter = new TypeAnalyzer();
            var result = reporter.Analyze(file);
            Approvals.VerifyAll(result, x => String.Format("{0},{1},{2},{3},{4}",
                x.Namespace,
                x.Name,
                x.FileDirectory,
                x.FileName,
                x.FileVersion
                ));
        }

        [Test]
        public void CanGenerateTypeReportForListOfAssemblies()
        {
            var files = Directory.GetFiles(".", "*.dll", SearchOption.AllDirectories)
                .Take(3)
                .Select(f => new FileInfo(f))
                .ToArray();

            var reporter = new TypeAnalyzer();

            var result = reporter.Analyze(files);

            Approvals.VerifyAll(result, x => String.Format("{0},{1},{2},{3},{4}",
                x.Namespace,
                x.Name,
                x.FileDirectory,
                x.FileName,
                x.FileVersion
                ));
        }

        [Test]
        public void CanGenerateNamespaceReport()
        {
            var asm = @".\nunit.framework.dll";
            var file = new FileInfo(asm);
            var reporter = new TypeAnalyzer();
            var result = reporter.SummarizeNamespaces(file);
            Approvals.VerifyAll(result, x => String.Format("{0},{1},{2},{3}",
                x.AssemblyName,
                x.AssemblyVersion,
                x.Namespace,
                x.Count
                ));
        }

        [Test]
        public void CanGenerateNamespaceReportForListOfAssemblies()
        {
            var files = Directory.GetFiles(".", "*.dll", SearchOption.AllDirectories)
                .Take(3)
                .Select(f => new FileInfo(f))
                .ToArray();

            var reporter = new TypeAnalyzer();

            var result = reporter.SummarizeNamespaces(files);

            Approvals.VerifyAll(result, x => String.Format("{0},{1},{2},{3}",
                x.Namespace,
                x.Count,
                x.AssemblyName,
                x.AssemblyVersion
                ));
        }
    }
}
