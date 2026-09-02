[![Build and Test](https://github.com/QIQ-Engine/QIQ/actions/workflows/build-and-test.yml/badge.svg)](https://github.com/QIQ-Engine/QIQ/actions/workflows/build-and-test.yml)
# QIQ <img style="height: 1em;" src="doc/Rabbit.svg">

QIQ (**Q**uick **I**nterpreter for **Q**uasi-PHP, /kik/) is an implementation of the [PHP language specification](https://phplang.org/) written in the Go programming language.

The goals of the project are:
- Deep dive into the PHP language syntax and the internals of its mode of operation
- Gain more experience in writing lexers, parser and interpreter
- *Very long-term goal*: Implement as many parts of the standard library and language features as needed to run a simple Laravel application :sweat:

**More documentation:**
- [AST](doc/AST.md)
- [Features](doc/Features.md)
- [Internal workings](doc/Internal%20workings.md)
- [Packages](doc/Packages.md)
- [Development](doc/Development.md)

## Usage
```
Usage of ./qiq:
  -h                  Show help.
  -debug              Enable debug mode (for development on QIQ).
  -dev                Run in developer mode.
  -stats              Show statistics.
  -S string           Run with built-in web server. <addr>:<port>
  -t string           Specify document root <docroot> for built-in web server.
  -f string           Parse and execute <file>.
  -ini string         Path to ini file to use.
  -create-ini string  Create given ini file.
```

**Parse file:**  
`cat index.php | ./qiq` or `./qiq -f index.php`

**Run web server:**  
`./qiq -S localhost:8080` - Document root is current working directory  
`./qiq -S localhost:8080 -dev` - Web server in developer mode  
`./qiq -S localhost:8080 -t /srv/www/html` - Document root is `/srv/www/html`

## Run official PHP `phpt` test cases
There are a lot of test cases in the source repository for PHP under the folder [tests](https://github.com/php/php-src/tree/master/tests).  
In order to test the QIQ implementation against this cases the binary `qiqTester` can be used.

**Usage:**
```bash
./qiqTester [-h] [-v(1|2)] [-only-failed] [-no-color] [-replace-json <file-name>] <list of directory or phpt-file>
```

**Examples:**
```bash
./qiqTester php-src/tests
./qiqTester -v2 php-src/tests/basic/001.phpt
```

## Usage with Docker
If you want to test or use QIQ with Docker, we've got you covered!

You can use the latest version: (This is not recommended as it might by unstable):
```bash
docker pull ghcr.io/qiq-engine/qiq:latest
```

Or use a specific version:
```bash
docker pull ghcr.io/qiq-engine/qiq:v0.4.0
```

You can find all versions [here](https://github.com/MasterZydra/QIQ/pkgs/container/qiq/versions).

### Run the docker image 
```bash
docker run -p 8080:8080 ghcr.io/qiq-engine/qiq:latest
```

You can change the port used inside the container (default: *8080*):
```bash
docker run -p 8081:8081 --env PORT=8081 ghcr.io/qiq-engine/qiq:latest
```

You can change the document root (default: */var/www/html*)
```bash
docker run -p 8080:8080 --env DOC_ROOT=/var/www/html/public ghcr.io/qiq-engine/qiq:latest
```

You can run the QIQ server in development mode (default: *false*)
```bash
docker run -p 8080:8080 --env DEV=true ghcr.io/qiq-engine/qiq:latest
```

You can also mount a local project into the container:
```bash
docker run -p 8080:8080 -v $(pwd):/var/www/html:z ghcr.io/qiq-engine/qiq:latest
```

## Special features

In addition to supporting official PHP syntax and features, QIQ offers some unique capabilities:

### Case Sensitive Files Include/Require

QIQ provides the `qiq.case_sensitive_include` INI directive, which enforces case-sensitive behavior for `include` and `require` statements even on Windows systems. This helps catch issues where file or path casing does not match, which would otherwise only cause problems in Linux environments.

### Strict Comparison Mode

QIQ provides the `qiq.strict_comparison` INI directive, which enforces strict comparison semantics throughout your codebase. When enabled, the `==` operator behaves like `===`, and both `!=` and `<>` behave like `!==`. This can help to prevent subtle bugs caused by type juggling in comparisons, making the code more predictable and robust.

## INI files

The application loads configuration files in the following order:
1. The INI file `qiq.ini` located in the same directory as the executable.
2. The INI file `qiq.ini` located in the document root.
3. Any custom INI file specified by the user using the `-ini` flag.

You can generate a new INI file using the `-create-ini` flag:

```bash
qiq -create-ini <path>/qiq.ini
```

## Used resources
For some part of this project, the following resources were used as a guide, inspiration, or concept:
- [PHP Language Specification](https://phplang.org/)
- YouTube playlist [Build a Custom Scripting Language In Typescript](https://www.youtube.com/playlist?list=PL_2VhOvlMk4UHGqYCLWc6GO8FaPl8fQTh) by [tylerlaceby](https://www.youtube.com/@tylerlaceby)
- Book [Crafting Interpreters](https://craftinginterpreters.com/) by Robert Nystorm
- Book [Writing An Interpreter In Go](https://interpreterbook.com/) by Thorsten Ball
