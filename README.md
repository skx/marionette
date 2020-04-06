# marionette

This is a proof of concept application which is designed to be puppet-like,
allowing you to define rules to get a system into a particular state.

At the moment there is support for:

* Defining rules.
  * Rules can have dependencies.
  * Rules can notify other rules when they're executed.
* Executing rules.


# Rule Definition

The general form of this a rule requires a module to be specified, along
with some module-specific parameters:

```
   $MODULE [triggered] {
              name => "NAME OF RULE",
              arg_1 => "Value 1 ... ",
              arg_2 => [ "array values", "are fine" ],
              arg_3 => "Value 3 .. ",
   }
```

In addition to the general arguments you can also specify dependencies via two magical keys:

* `dependencies`
  * A list of any number of rules which must be executed before this.
* `notify`
  * A list of any number of rules which should be notified, because this rule was triggered.
    * Triggered in this sense means that the rule was executed and the state changed.

As a concrete example you need to run a command which depends upon a directory being present.  You could do this like so:


```
shell{ name         => "Test",
       command      => "uptime > /tmp/blah/uptime",
       dependencies => [ "Create /tmp/blah" ] }

directory{ name   => "Create /tmp/blah",
           target => "/tmp/blah" }
```

The alternative would have been to have the directory creation trigger the other rule via notification:

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


# Primitive Types

We only have a small number of primitives at the moment, however the dependency resolution and notification system is reliable.


## `file`

The file module allows a file to be created, from a local file, or via a remote HTTP-source.

Example usage:

```
file {  name       => "fetch file",
        target     => "/tmp/steve.txt",
        source_url => "https://steve.fi/",
}
```

`target` is a mandatory parameter, and specifies the file to be operated upon.

There are two ways a file can be created:

* `source_url` - Fetched from the remote URL.
* `source` - Copied from the existing path.

Other valid parameters are:

* `owner` - Username of the owner, e.g. "root".
* `group` - Groupname of the owner, e.g. "root".
* `mode` - THe mode to set, e.g. "0755".


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


## `shell`

The shell module allows you to run shell-commands.

Example:

```
shell { name => "I touch your file.",
        command => "touch /tmp/blah/test.me" }
```

`command` is the only mandatory parameter.


# Example

See [input.txt](input.txt) for a sample rule-file, including syntax breakdown.
