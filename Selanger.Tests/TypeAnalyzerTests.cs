using System;
using System.IO;
using ApprovalTests;
using NUnit.Framework;
using Selanger.Core;

namespace Selanger.Tests
{
    [TestFixture]
    public class TypeAnalyzerTests
    {
        [Test]
        public void Throws_if_file_not_found()
        {
            var asm = @".\i.do.not.exist.dll";
            var file = new FileInfo(asm);
            var reporter = new TypeAnalyzer();

            var ex = Assert.Throws<InvalidOperationException>(() =>
                {
                    reporter.Analyze(file);
                });

            Assert.AreEqual("Could not find", ex.Message.Substring(0,14));
        }

        [Test]
        public void CanGenerateTypeReport()
        {
            var asm = @".\nunit.framework.dll";
            var file = new FileInfo(asm);
            var reporter = new TypeAnalyzer();
            var result = reporter.Analyze(file);
            Approvals.VerifyAll(result, x => String.Format("{0},{1},{2},{3}", 
                                                            x.AssemblyName, 
                                                            x.AssemblyVersion,
                                                            x.Namespace, 
                                                            x.Count
                                                           ));
        }
    }
}
