[![Go Report Card](https://goreportcard.com/badge/github.com/skx/marionette)](https://goreportcard.com/report/github.com/skx/marionette)
[![license](https://img.shields.io/github/license/skx/marionette.svg)](https://github.com/skx/marionette/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/skx/marionette.svg)](https://github.com/skx/marionette/releases/latest)

* [marionette](#marionette)
* [Installation &amp; Usage](#installation--usage)
* [Rule Definition](#rule-definition)
* [Module Types](#module-types)
   * [directory](#directory)
   * [dpkg](#dpkg)
   * [edit](#edit)
   * [file](#file)
   * [git](#git)
   * [link](#link)
   * [shell](#shell)
* [Adding Modules](#adding-modules)
* [Example Rules](#example-rules)
* [Future Plans](#future-plans)
* [Github Setup](#github-setup)


# marionette

This is a proof of concept application which is designed to be puppet-like,
allowing you to define rules to get a system into a particular state.

We have built-in support for the minimum set of primitives you need to turn a blank virtual machine into a host running a few services:

* Cloning git repositories.
* Creating/modifying files/directories.
* Removing packages.
* Triggering shell actions.

The available modules for performing tasks can be extended via the use of [external plugins](#adding-modules) so you don't necessarily need to be a Golang coder to add new things.




# Installation & Usage

Binaries for several systems are available upon our [download page](https://github.com/skx/marionette/releases).  If you prefer to use something more recent you can install directly from the repository via:

```
go get github.com/skx/marionette
```

Once installed you can then execute it with the path of one or more rule-files like so:

```
marionette ./rules.txt ./rules2.txt ... ./rulesN.txt
```




# Rule Definition

The general form of our rules looks like this:

```
$MODULE [triggered] {
            name => "NAME OF RULE",
            arg_1 => "Value 1 ... ",
            arg_2 => [ "array values", "are fine" ],
            arg_3 => "Value 3 .. ",
}
```

Each rule starts by declaring the type of module which is being invoked, then
there is a block containing "`key => value`" sections.  Different modules will
accept and expect different keys to configure themselves.  (Unknown arguments
will be ignored.)

Do note that every rule needs a name, and that must be unique.  Names are used
to trigger/define dependencies.  In the (near) future this will be optional.

In addition to the general arguments passed to the available modules you can also specify dependencies via two magical keys within each rule block:

* `dependencies`
  * A list of any number of rules which must be executed before this.
* `notify`
  * A list of any number of rules which should be notified, because this rule was triggered.
    * _Triggered_ in this sense means that the rule was executed and the state changed.

As a concrete example you need to run a command which depends upon a directory being present.  You could do this like so:


```
shell { name         => "Test",
        command      => "uptime > /tmp/blah/uptime",
        dependencies => [ "Create /tmp/blah" ] }

directory{ name   => "Create /tmp/blah",
           target => "/tmp/blah" }
```

The alternative would have been to have the directory-creation trigger the shell-execution rule via an explicit notification:

```
directory{ name   => "Create /tmp/blah",
           target => "/tmp/blah",
           notify => [ "Test" ]
}

shell triggered { name         => "Test",
                  command      => "uptime > /tmp/blah/uptime",
}
```

The difference in these two approaches is how often things run:

* In the first case we always run `uptime > /tmp/blah/uptime`
  * We just make sure that _before_ that the directory has been created.
* In the second case we run the command only once.
  * We run it only after the directory is created.
  * Because the directory-creation triggers the notification only when the rule changes (i.e. the directory goes from being "absent" to "present").

You'll note that any module which is followed by the token `triggered` will __only__ be executed when it is triggered by name.  If there is no `notify` key referring to that rule it will __never__ be executed.




# Module Types

We have a small number of primitives at the moment implemented in 100% pure golang.  Adding [new modules as plugins](#adding-modules) is possible, and contributions for various purposes are most welcome.



## `directory`

The directory module allows you to create a directory, or change the permissions of one.

Example usage:

```
directory {  name    => "My home should have a binary directory",
             target  => "/home/steve/bin",
             mode    => "0755",
}
```

Valid parameters are:

* `target` is a mandatory parameter, and specifies the directory to be operated upon.
* `owner` - Username of the owner, e.g. "root".
* `group` - Groupname of the owner, e.g. "root".
* `mode` - The mode to set, e.g. "0755".
* `state` - Set the state of the directory.
  * `state => "absent"` remove it.
  * `state => "present"` create it (this is the default).



## `dpkg`

This module allows purging a package, or set of packages:

```
dpkg { name => "Remove stuff",
       package => ["vlc", "vlc-l10n"] }
```

Only the `package` key is required.

In the future we _might_ have an `apt` module for installing new packages.  We'll see.



## `edit`

This module allows minor edits to be applied to a file:

* Removing lines matching a given regular expression.
* Appending a line to the file if missing.

```
edit { name => "Remove my VPN hosts",
       target => "/etc/hosts",
       remove_lines => "\.vpn" }
```

The following keys are supported:

* `target` - Mandatory filename to edit.
* `remove_lines` - Remove any lines of the file matching the specified regular expression.
* `append_if_missing` - Append the given text if not already present in the file.



## `file`

The file module allows a file to be created, from a local file, or via a remote HTTP-source.

Example usage:

```
file {  name       => "fetch file",
        target     => "/tmp/steve.txt",
        source_url => "https://steve.fi/",
}

file {  name     => "write file",
        target   => "/tmp/name.txt",
        content  => "My name is Steve",
}
```

`target` is a mandatory parameter, and specifies the file to be operated upon.

There are three ways a file can be created:

* `content` - Specify the content inline.
* `source_url` - The file contents are fetched from a remote URL.
* `source` - Content is copied from the existing path.

Other valid parameters are:

* `owner` - Username of the owner, e.g. "root".
* `group` - Groupname of the owner, e.g. "root".
* `mode` - The mode to set, e.g. "0755".
* `state` - Set the state of the file.
  * `state => "absent"` remove it.
  * `state => "present"` create it (this is the default).



## `git`

Clone a remote repository to a local directory.

Example usage:

```
git { path => "/tmp/xxx",
      repository => "https://github.com/src-d/go-git",
}
```

Valid parameters are:

* `repository` Contain the HTTP/HTTPS repository to clone.
* `path` - The location we'll clone to.
* `branch` - The branch to switch to, or be upon.
  * A missing branch will not be created.

If this module is used to `notify` another then it will trigger such a
notification if either:

* The repository wasn't present, and had to be cloned.
* The repository was updated.  (i.e. Remote changes were pulled in.)



## `link`

The `link` module allows you to create a symbolic link.

Example usage:

```
link { name => "Symlink test",
       source => "/etc/passwd",  target => "/tmp/password.txt" }
```

Valid parameters are:

* `target` is a mandatory parameter, and specifies the location of the symlink to create.
* `source` is a mandatory parameter, and specifies the item the symlink should point to.



## `shell`

The shell module allows you to run shell-commands.

Example:

```
shell { name => "I touch your file.",
        command => "touch /tmp/blah/test.me" }
```

`command` is the only mandatory parameter.




# Adding Modules

Adding marionette modules can be done in two ways:

* Writing your module in 100% pure go.
  * This is how the bundled modules are written, there is a simple API which only requires implementing the methods in the [ModuleAPI interface](modules/api.go).
* Writing an external/binary  plugin.
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



# Example Rules

There is an example ruleset included in the distribution:

* [input.txt](input.txt)

That should be safe to run for all users, as it only modifies files beneath `/tmp`.




# Future Plans

* I think it would be nice to include some binary plugins embedded within our application, via [implant](https://github.com/skx/implant).  That would allow fast development and avoid the need for local filesystem access to execute them.
* We need to support installing packages upon a Debian GNU/Linux host, not just purging unwanted ones.
* Gathering "facts" about the local system, and storing them as variables would be useful.
  * Of course then we might want to do conditional actions.
  * I guess we could write something like this:
    * `only_if => "${distribution} == Debian"`
    * But of course that gets messy.  Compare with Ansible, again.




## Github Setup

This repository is configured to run tests upon every commit, and when
pull-requests are created/updated.  The testing is carried out via
[.github/run-tests.sh](.github/run-tests.sh) which is used by the
[github-action-tester](https://github.com/skx/github-action-tester) action.

Releases are automated in a similar fashion via [.github/build](.github/build),
and the [github-action-publish-binaries](https://github.com/skx/github-action-publish-binaries) action.


Steve
--
