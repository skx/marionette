# marionette

This is a proof of concept application which is designed to be puppet-like,
allowing you to define rules to get a system into a particular state.

At the moment there is support for:

* Defining rules.
  * Rules can have dependencies.
  * Rules can notify other rules when they're executed.
* Executing rules.

# Primitive Types

We only have a small number of primitives at the moment, however the dependency resolution and notification system is reliable

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


## `directory`

The directory module allows you to create a directory, or change the permissions of one.

Example usage:

```
directory {  name    => "My home should have a binary directory",
             target  => "/home/steve/bin",
             mode    => "0755",
}
```

`target` is a mandatory parameter, and specifies the directory to be operated upon.
`mode` is optional and sets the mode.



## `shell`

The shell module allows you to run shell-commands.

Example:

```
shell { name => "I touch your file.",
        command => "touch /tmp/blah/test.me" }
```

`command` is a mandatory parameter.


# Example

See [input.txt](input.txt) for a sample rule-file, including syntax breakdown.
