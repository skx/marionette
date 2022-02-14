[![Go Report Card](https://goreportcard.com/badge/github.com/skx/marionette)](https://goreportcard.com/report/github.com/skx/marionette)
[![license](https://img.shields.io/github/license/skx/marionette.svg)](https://github.com/skx/marionette/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/skx/marionette.svg)](https://github.com/skx/marionette/releases/latest)

* [marionette](#marionette)
* [Installation &amp; Usage](#installation--usage)
* [Rule Definition](#rule-definition)
  * [Dependency Management](#dependency-management)
  * [Conditionals](#conditionals)
  * [Examples](#examples)
  * [Misc. Features](#misc-features)
    * [Command Execution](#command-execution)
    * [File Inclusion](#include-files)
    * [Pre-Declared Variables](#pre-declared-variables)
* [Module Types](#module-types)
   * [directory](#directory)
   * [docker](#docker)
   * [edit](#edit)
   * [file](#file)
   * [git](#git)
   * [group](#group)
   * [http](#http)
   * [link](#link)
   * [log](#log)
   * [package](#package)
   * [shell](#shell)
   * [sql](#sql)
   * [user](#user)
* [Future Plans](#future-plans)
  * [See also](#see-also)
* [Github Setup](#github-setup)


# marionette

`marionette` is a simple command-line application which is designed to carry out system automation tasks.  It was designed to resemble the well-known configuration-management application [puppet](https://puppet.com/).

`marionette` contains a small number of built-in modules, providing enough primitives to turn a blank virtual machine into a host running a few services:

* Cloning git repositories.
* Creating/modifying files/directories.
* Fetching Docker images.
* Installing and removing packages.
  * Debian GNU/Linux, and CentOS are supported, using `apt-get`, `dpkg`, and `yum` as appropriate..
* Executing shell commands.

In the future it is possible that more modules will be added, but this will require users to file bug-reports requesting them, contribute code, or the author realizing something is necessary.




# Installation & Usage

Binaries for several systems are available upon our [download page](https://github.com/skx/marionette/releases).  If you prefer to use something more recent you can install directly from the repository via:

```
go install github.com/skx/marionette@latest
```

The main application can then be launched with the path to a set of rules, which it will then try to apply:

```
marionette [flags] ./rules.txt ./rules2.txt ... ./rulesN.txt
```

Currently there are two supported flags:

* `-verbose`
  * Show extra details when executing the supplied rules-file(s).
* `-debug`
  * Show many low-level details when executing the supplied rules-file(s).

Typically a user would run with `-verbose`, and a developer might examine the output produced when `-debug` is specified.



# Rule Definition

The general form of our rules looks like this:

```
$MODULE [triggered] {
            name  => "NAME OF RULE",
            arg_1 => "Value 1 ... ",
            arg_2 => [ "array values", "are fine" ],
            arg_3 => "Value 3 .. ",
}
```

Each rule starts by declaring the type of module which is being invoked, then there is a block containing "`key => value`" sections.  Different modules will accept/expect different keys to configure themselves.  (Unknown arguments will generally be ignored.)

A rule may also contain an optional `triggered` attribute.  Rules which contain the `triggered` modifier are not executed unless explicitly invoked by another rule - think of it as a "handler" if you're used to `ansible`.

Here is an example rule which executes a shell-command:

```
# Run a command, unconditionally
shell {
        command => "uptime > /tmp/uptime.txt"
}

```

Another simple example to illustrate the available syntax might look like the following, which ensures that I have `~/bin/` owned by myself:

```
directory {
             target  => "/home/${USER}/bin",
             mode    => "0755",
             owner   => "${USER}",
             group   => "${USER}",
}
```

There are four magical keys which can be supplied to all modules:

| Name     | Usage                                                            |
|----------|------------------------------------------------------------------|
| `require`| This is used for [dependency management](#dependency-management) |
| `notify` | This is used for [dependency management](#dependency-management) |
| `if`     | This is used to make a rule [conditional](#conditionals)         |
| `unless` | This is used to make a rule [conditional](#conditionals)         |


## Dependency Management

There are two keys which can be used to link rules together, to handle dependencies:

* `require`
  * This key contains either a single rule-name, or a list of any rule-names, which must be executed before _this_ one.
* `notify`
  * A list of any number of rules which should be notified, if the given rule resulted in a state-change.

**Note** You only need to give rules names to link them for the purpose of managing dependencies.

Imagine we wanted to create a new directory, and write a file there.  We could do that with a pair of rules:

* One to create a directory.
* One to generate the output file.

We could wing-it and write the rules in the logical order, but it would be far better to link the two rules explicitly.

There are two ways we could implement this.  The simplest way would be this:

```
shell { command   => "uptime > /tmp/blah/uptime",
        require   => "Create /tmp/blah" }

directory{ name   => "Create /tmp/blah",
           target => "/tmp/blah" }
```


The alternative would have been to have the directory-creation trigger the shell-execution rule via an explicit notification:

```
# This command will notify the "Test" rule, if it creates the directory
# because it was not already present.
directory{ target => "/tmp/blah",
           notify => "Test"
}

# Run the command, when triggered/notified.
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

You'll note that any rule which is followed by the token `triggered` will __only__ be executed when it is triggered by name.  If there is no `notify` key referring to that rule it will __never__ be executed.




## Conditionals

Rules may be made conditional, via the magical keys `if` and `unless`.

The following example runs a command, using `apt-get`, only if a specific file exists upon the filesystem:

```
shell { name    => "Upgrade",
        command => "apt-get dist-upgrade --yes --force-yes",
        if      => exists(/usr/bin/apt-get) }
```

For the reverse, running a rule unless something is true, we can use the `unless` key:

```
let arch = `/usr/bin/arch`

file { name   => "Create file",
       target => "/tmp/foo",
       unless => equal( "x86_64", "${arch}" ) }
```

Here we see that we've used two functions `equal` and `exists`, these are both built-in functions which do what you'd expect.

The following list shows all the built-in functions that you may use (but only within the `if` or `unless` keys):

* `contains(haystack, needle)`
  * Returns true if the first string contains the second.
* `exists( /some/file )`
  * Return true if the specified file/directory exists.
* `equal( foo, bar )`
  * Return true if the two values are identical.
* `nonempty(string|variable)`
  * Return true if the string/variable is non-empty.
  * `set` is a synonym.
* `on_path(string|variable)`
  * Return true if a binary with the given name is on the users' PATH.
* `empty(string|variable)`
  * Return true if the string/variable is empty (i.e. has zero length).
  * `unset` is a synonym.
* `success(string)`
  * Returns true if the command `string` is executed and returns a non-error exit-code (i.e. 0).
  * Output is discarded, and not captured.
* `failure(string)`
  * Returns true if the command `string` is executed and returns an error exit-code (i.e. non-zero 0).
  * Output is discarded, and not captured.

More conditional primitives may be added if they appear to be necessary, or if users request them.

**NOTE**: The conditionals are only supported when present in keys named `if` or `unless`.  This syntax is special for those two key-types, however conditionals may also be applied to variable assignments and file inclusion:

```
# Include a file of rules, on a per-arch basis
include "x86_64.rules" if equal( "${ARCH}","x86_64" )
include "i386.rules"   if equal( "${ARCH}","i386" )

# Setup a ${cmd} to download something, depending on what is present.
let cmd = "curl --output ${dst} ${url}" if on_path("curl")
let cmd = "wget -O ${dst} ${url}"       if on_path("wget")
```


## Examples

You can find a small set of example recipes beneath the examples directory:

* [examples/](examples/)



## Misc. Features


### Command Execution

Backticks can be used to execute commands, in variable-assignments and in parameters to rules.

For example we might determine the system architecture like this:

```
let arch = `/usr/bin/arch`

shell { name    => "Show arch",
        command => "echo We are running on an ${arch} system" }
```

Here `${arch}` expands to the output of the command, as you would expect, with any trailing newline removed.

> **Note** `${ARCH}` is available by default, as noted in the [pre-declared variables](#pre-declared-variables) section.  This was just an example of command-execution.

Using commands inside parameter values is also supported:

```
file { name    => "set-todays-date",
       target  => "/tmp/today",
       content => `/usr/bin/date` }
```

The commands executed with the backticks have any embedded variables expanded _before_ they run, so this works as you'd expect:

```
let fmt   = "+%Y"

file { name    => "set-todays-date",
       target  => "/tmp/today",
       content => `/bin/date ${fmt}` }
```


### Include Files

You can break large rule-files into pieces, and include them in each other:

```
# main.in

let prefix="/etc/marionette"

include "foo.in"
include "${prefix}/test.in"
```

To simplify your recipe writing including other files may be made conditional,
just like our rules:

```
# main.in

include "x86_64.rules" if equal( "${ARCH}","x86_64" )
include "i386.rules"   if equal( "${ARCH}","i386" )
```


### Pre-Declared Variables

The following variables are available by default:

| Name          | Value                                                 |
|---------------|-------------------------------------------------------|
| `${ARCH}`     | The system architecture (as taken from `sys.GOARCH`). |
| `${HOMEDIR}`  | The home directory of the user running marionette.    |
| `${HOSTNAME}` | The hostname of the local system.                     |
| `${OS}`       | The operating system name (as taken from `sys.GOOS`). |
| `${USERNAME}` | The username of user running marionette.              |

There are additionally two "magic" variables available which will always have values based upon the current rule-file being processed, whether that is a file specified upon the command-line, or as a result of an `include` statement:

| Name              | Value                                                            |
|-------------------|------------------------------------------------------------------|
| `${INCLUDE_DIR}`  | The absolute directory path of the current file being processed. |
| `${INCLUDE_FILE}` | The absolute path of the current file being processed.           |



# Module Types

Our primitives are implemented in 100% pure golang, and are included with our binary, these are now described briefly:


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

* `target` is a mandatory parameter, and specifies the directory, or directories, to operate upon.
* `owner` - Username of the owner, e.g. "root".
* `group` - Groupname of the owner, e.g. "root".
* `mode` - The mode to set, e.g. "0755".
* `state` - Set the state of the directory.
  * `state => "absent"` remove it.
  * `state => "present"` create it (this is the default).



## `docker`

This module allows fetching a container from a remote registry.

```
docker { image => "alpine:latest" }
```

The following keys are supported:

* `image` - The image/images to fetch.
* `force`
  * If this is set to `true` then we fetch the image even if it appears to be available locally already.

**NOTE**: We don't support private registries, or the use of authentication.



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
* `search`
* `replace`
  * If both `search` and `replace` are non-empty then they will be used to update the content of the specified file.
  * `search` is treated as a regular expression, for added flexibility.

An example of changing a file might look like this:

```
edit { target  => "/etc/ssh/sshd_config",
       search  => "PasswordAuthentication",
       replace => "# PasswordAuthentication",
}
```


## `fail`

The fail-module is designed to terminate processing, if you find a situation where the local
environment doesn't match your requirements.  For example:

```
let path = `which useradd`

fail {
   message => "I can't find a working useradd binary to use",
   if      => empty(path)
}
```

The only valid parameter is `message`, which must be a single string.

See also [log](#log), which will log a message but then continue execution.



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

There are four ways a file can be created:

* `content` - Specify the content inline.
* `source_url` - The file contents are fetched from a remote URL.
* `source` - Content is copied from the existing path.
* `template` - Content is produced by rendering a template from a path.

Other valid parameters are:

* `owner` - Username of the owner, e.g. "root".
* `group` - Groupname of the owner, e.g. "root".
* `mode` - The mode to set, e.g. "0755".
* `state` - Set the state of the file.
  * `state => "absent"` remove it.
  * `state => "present"` create it (this is the default).

Where `template` is used, the template file is rendered using the
[`text/template`](https://pkg.go.dev/text/template) Go package


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



## `group`

The group module allows you to add or remove local Unix groups to your system.

Example:

```
group { group => "sysadmin",
        state => "present" }
```

* `group` is a mandatory parameter.
* `state` should be one of `absent` or `present`, depending upon whether you want to add or remove the group.



## `http`

The `http` module allows you to make HTTP requests.

Example:

```
http { url => "https://api.github.com/user",
       method => "PATCH",
       headers => [
         "Authorization: token ${AUTH_TOKEN}",
         "Accept: application/vnd.github.v3+json",
       ],
       body => "{\"name\":\"new name\"}",
       expect => "200" }
```

Valid parameters are:

* `url` is a mandatory parameter, and specifies the URL to make the HTTP request to.
* `method` - If this is set, the HTTP request will use this method, otherwise a `GET` request is made.
* `headers` - If this is set, the headers provided will be sent with the request.
* `body` - If this is set, this will be the body of the HTTP request.
* `expect` - If this is set, an error will be triggered if the response status code does not match the expected status code.
  * If `expect` is not set, an error will be triggered for any non 2xx response status code.

The `http` module is always regarded as having made a change on a successful request.

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



## `log`

The `log` module allows you to output a message, but continue execution.

Example usage:

```
log { message => "I'm ${USER} running on ${HOSTNAME}" }
```

The only valid parameter is `message`, which may be either a single value, or an array of values.

See also [fail](#fail), which will log a message but then terminate execution.



## `package`

The package-module allows you to install or remove a package from your system, via the execution of `apt-get`, `dpkg`, and `yum`, as appropriate.

Example usage:

```
# Install a single package
package { name    => "Install bash",
          package => "bash",
          state   => "installed",
        }

# Uninstall a series of packages
package { package => [ "nano", "vim-tiny", "nvi" ],
          state => "absent" }
```

Valid parameters are:

* `package` is a mandatory parameter, containing the package, or list of packages.
* `state` - Should be one of `installed` or `absent`, depending upon whether you want to install or uninstall the named package(s).
* `update` - If this is set to `true` then the system will be updated prior to installation.
  * In the case of a Debian system, for example, `apt-get update` will be executed.



## `sql`

The SQL-module allows you to run arbitrary SQL against a database.  Two parameters are required `driver` and `dsn`, which are used to open the database connection.  For example:

```
sql {
       driver => "sqlite3",
       dsn    => "/tmp/sql.db",

       # Memory based-SQLite could be used like so:
       #  dsn   => "/tmp/foo.db?mode=memory",

       # MySQL connection would look like this:
       #
       #   driver => "mysql",
       #   dsn    => "user:password@(127.0.0.1:3306)/",
}
```

We support the following drivers:

* `mysql`
* `postgres`
* `sqlite3`

To specify the query to run you should set one of the following two parameters:

* `sql`
  * Literal SQL to execute.
* `sql_file`
  * A file to read and execute in one execution.

NOTE: You may find you need to append `multiStatements=true` to your DSN to ensure correct operation when reading SQL from a file.


## `shell`

The shell module allows you to run shell-commands, complete with redirection and pipes.

Example:

```
shell { name    => "I touch your file.",
        command => "touch /tmp/blah/test.me"
      }
```

`command` is the only mandatory parameter.  Multiple values can be specified:

```
shell { name    => "restart the aplication",
        command => [
                     "systemctl stop  foo.service",
                     "systemctl start foo.service",
                   ]
      }
```

By default commands are executed directly, unless they contain redirection-characters (">", or "<"), or the use of a pipe ("|").  If special characters are used then we instead invoke the command via `/bin/bash`:

* `bash -c "${command}"`

You may specify `shell => true` to force the use of a shell, despite the lack of redirection/pipe characters:

```
shell { shell   => true,
        command => "sed .. /etc/file.txt"
      }
```


## `user`

The user module allows you to add or remove local Unix users to your system.

Example:

```
user { login => "steve",
       state => "present" }
```

* `login` is a mandatory parameter.
* `shell` is an optional parameter to use for the users' shell.
* `state` should be one of `absent` or `present`, depending upon whether you want to add or remove the user.



# Future Plans

* Gathering "facts" about the local system, and storing them as variables would be useful.
  * At the moment we just have a small number of [pre-declared variables](#pre-declared-variables).


## See Also

* There are some brief notes on our implementation contained within [HACKING.md](HACKING.md).
  * This gives an overview of the structure.
* There are some brief-notes on [fuzz-testing the parser](FUZZING.md).
  * This should ensure that there is no input that can crash our application.


## Github Setup

This repository is configured to run tests upon every commit, and when
pull-requests are created/updated.  The testing is carried out via
[.github/run-tests.sh](.github/run-tests.sh) which is used by the
[github-action-tester](https://github.com/skx/github-action-tester) action.

Releases are automated in a similar fashion via [.github/build](.github/build),
and the [github-action-publish-binaries](https://github.com/skx/github-action-publish-binaries) action.


Steve
--
