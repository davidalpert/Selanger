using System;
using System.IO;

namespace Selanger.CLI.Tests.Helpers
{
    public class RedirectedConsole : IDisposable
    {
        readonly TextWriter _outBefore;
        readonly TextWriter _errBefore;
        readonly StringWriter _console;
 
        public RedirectedConsole()
        {
            _console = new StringWriter();
            _outBefore = Console.Out;
            _errBefore = Console.Error;
            Console.SetOut(_console);
            Console.SetError(_console);
        }
 
        public string Output
        {
            get { return _console.ToString(); }
        }
 
        public void Dispose()
        {
            Console.SetOut(_outBefore);
            Console.SetError(_errBefore);
            _console.Dispose();
        }
    }
}