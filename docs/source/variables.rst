Config Variables
================

Some fields in ``dobi.yaml`` support variable interpolation from values provided
by **dobi** or the shell environment.  Variables are wrapped in braces, for example
``{env.USER}`` would inject the value of the ``$USER`` environment variable.

Format
------

Environment variables have the following format:


.. code-block:: default

    "{" [section.]variable[:default] "}"

**{}**
    All variables are wrapped in braces

**section**
    Some variables are grouped into sections (like **git** or **env**)

**default**
    Variables can have default values. The value after the last colon is taken
    as the default value. An empty default value makes the variable act like an
    optional variable.

Example
~~~~~~~

Use the value of ``$VERSION`` from the host use the variable:

.. code-block:: none

    {env.VERSION}

Use a default value of ``v1.0``:

.. code-block:: none

    {env.VERSION:v1.0}

Use a variable with an empty default value as an optional value:

.. code-blcok:: none

    {env.VERSION:}


Supported Variables
-------------------

The following variables are made avariables:

* ``env.<variable>`` - the value of an environment variable
* ``git.sha`` - the current git sha
* ``git.short-sha`` - the first 10 characters of the current git sha
* ``git.branch`` - the current git branch name
* ``time.<format>`` - a date or time using `fmtdate
  <https://github.com/metakeule/fmtdate#placeholders>`_ (note: if your time
  format includes a ``:`` you must add another ``:`` to the end of the format,
  otherwise the string after the final ``:`` will be taken as the default value)
* ``unique`` - a unique execution id generate from the project name and exec id
* ``exec-id`` - an execution id (without project name)
* ``project`` - the project name


Config Fields
-------------

The following config fields support variables:

* ``run.env``
* ``run.net-mode``
* ``image.tag``
* ``image.args``
* ``compose.files``
* ``compose.project``
* ``meta.exec-id``
