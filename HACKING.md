
Some brief notes on implementation/internals.

* [Implementation Overview](#implementation-overview)




# Implementation Overview

Our implementation is pretty simple and all revolves around a set of rules.

* Half our code is involved with producing the rules:
  * We have a [lexer](lexer/) to split our input into a set of [tokens](token/).
  * The [parser](parser/) reads those tokens to convert an input-file into a series of [AST](ast/) objects.

* The other half of our code is involved with executing the rules.
  * The main driver is the [executor](executor/) package, which runs rules.
  * Conditional execution is managed via the [conditionals](conditionals/) package.

In addition to the above we also have a [config](config/) object which is passed around to allow us to centralize global state, and we have a set of [file](file/) helpers which contain some central code.


# Testing Overview

There is an associated github action to run our test-cases, and some linters, every time a pull-request is created/updated against the remote repository.

You should probably run the driver when you're testing:

     .github/run-tests.sh

Note that this installs some tools if the environmental variable "`$CI`" is set, so you might need to do that the first time:

     CI=true .github/run-tests.sh
