# Fuzz-Testing

The upcoming 1.18 release of the golang compiler/toolset has integrated
support for fuzz-testing.

Fuzz-testing is basically magical and involves generating new inputs "randomly"
and running test-cases with those inputs.


## Running

If you're running 1.18beta1 or higher you can run the fuzz-testing against
our  parser like so:

    $ cd parser/
    $ go test -fuzztime=300s -parallel=1 -fuzz=FuzzParser -v
    === RUN   TestBlock
    --- PASS: TestBlock (0.00s)
    === RUN   TestConditinalErrors
    --- PASS: TestConditinalErrors (0.00s)
    === RUN   TestConditional
    --- PASS: TestConditional (0.00s)
    === FUZZ  FuzzParser
    fuzz: elapsed: 0s, gathering baseline coverage: 0/149 completed
    fuzz: elapsed: 0s, gathering baseline coverage: 149/149 completed, now fuzzing with 1 workers
    fuzz: elapsed: 3s, execs: 42431 (14140/sec), new interesting: 0 (total: 143)
    fuzz: elapsed: 6s, execs: 93384 (16985/sec), new interesting: 0 (total: 143)
    fuzz: elapsed: 9s, execs: 145220 (17280/sec), new interesting: 0 (total: 143)
    fuzz: elapsed: 12s, execs: 193264 (16017/sec), new interesting: 0 (total: 143)
    ..
    fuzz: elapsed: 4m54s, execs: 5376429 (20034/sec), new interesting: 11 (total: 154)
    fuzz: elapsed: 4m57s, execs: 5436966 (20179/sec), new interesting: 11 (total: 154)
    fuzz: elapsed: 5m0s, execs: 5494052 (19027/sec), new interesting: 12 (total: 155)
    fuzz: elapsed: 5m1s, execs: 5494052 (0/sec), new interesting: 12 (total: 155)
    --- PASS: FuzzParser (301.02s)
    PASS
    ok  	github.com/skx/marionette/parser	301.042s


You'll note that I've added `-parellel=1` to the test, because otherwise my desktop system becomes unresponsive while the testing is going on.
