# Fuzz-Testing

If you don't have the appropriate tools installed you can fetch them via:

    $ go get github.com/dvyukov/go-fuzz/go-fuzz
    $ go get github.com/dvyukov/go-fuzz/go-fuzz-build

Now you can build fuzzing version of the parser:

    $ go-fuzz-build github.com/skx/marionette/fuzz

Create a location to hold the work, and give it copies of some sample-programs:

    $ mkdir -p workdir/corpus
    $ cp input.txt workdir/corpus

Now you can actually launch the fuzzer - here I use `-procs 1` so that
my desktop system isn't complete overloaded:

    $ export FUZZ=FUZZ
    $ go-fuzz -procs 1 -bin=fuzz-fuzz.zip -workdir workdir/

**NOTE**: We set the environmental variable `FUZZ` to ensure that system-commands are not executed.

Finally you'll see any crashing-programs in `workdir/crashers`.
