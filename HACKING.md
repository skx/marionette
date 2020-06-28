
Some brief notes on implementation/internals.

* [Implementation Overview](#implementation-overview)
* [Binary Plugins](#binary-plugins)




# Implementation Overview

Our implementation is pretty simple and all revolves around a set of rules.

* Half our code is involved with producing the rules:
  * We have a [lexer](lexer/) to split our input into a set of [tokens](token/).
  * The [parser](parser/) reads those tokens to convert an input-file into a series of [rules](rules/).
* The other half of our code is involved with executing the rules.
  * The main driver is the [executor](executor/) package, which runs rules.
  * Conditional execution is managed via the [conditionals](conditionals/) package.

In addition to the above we also have a [config](config/) object which is passed around to allow us to centralize global state, and we have a set of [file](file/) helpers which contain some central code.




# Binary Plugins

Adding marionette modules can be done in two ways:

* Writing your module in 100% pure go.
  * This is how the bundled modules are written, there is a simple API which only requires implementing the methods in the [ModuleAPI interface](modules/api.go).
* Writing an external/binary plugin.
  * You can write a plugin in __any__ language you like.

Given a rule such as the following we'll look for a handler for `foo`:

```
foo { name => "My rule",
      param1 => "One", }
```

If there is no built-in plugin with that name then instead we'll execute an external binary, if it exists.  The parameters supplied in the rule-block will be passed as JSON piped to STDIN when the process is launched.

If that plugin makes a change, such that triggers should be executed, it should print `changed` to STDOUT and exit with a return code of 0.  If no change was made then it should print `unchanged` to STDOUT and also exit with a return code of 0.

A non-zero return code will be assumed to mean something failed, and execution will terminate.

There are two directories searched for plugins:

* `/opt/marionette/plugins`
* `~/.marionette/plugins`

So in the example above we'd execute `/opt/marionette/plugins/foo` or `~/.marionette/plugins/foo` if they exist.  If there was no built-in module with the name, and no binary plugin found then we'd have to report an error and terminate.
