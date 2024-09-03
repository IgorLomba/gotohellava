<br />
<p align="center">

<h3 align="center">AVA-CLI</h3>

<p align="center">
    AVA CLI is a command-line tool designed to access ava.ufms.br website and visit all available links within that course.
<br />
</p>



<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li><a href="#getting-started">Getting Started</a></li>
    <li><a href="#pre-requisites">Pre-requisites</a></li>
    <li><a href="#compile">Compile</a></li>
    <li><a href="#usage">Usage</a></li>
  </ol>
</details>

## Getting Started

Read the document above for CLI commands and usage.

## Pre-requisites

### Go

To be able to build the code you should have:

* Go - You can download and install Go using this [link](https://golang.org/doc/install).

### Makefile

For ease of use, a Makefile is provided to build the project for all platforms.

Install Make for Mac: https://formulae.brew.sh/formula/make

Install Make for Windows: https://sourceforge.net/projects/gnuwin32/files/make/3.81/make-3.81.exe/download

## Compile

#### Windows

``` powershell
setx GOOS=windows 
setx GOARCH=amd64
go build -o ./bin/ava.exe .
```

#### Linux

``` bash
export GOARCH=amd64
export GOOS=linux
go build -o ./bin/ava .
```

#### Macintosh

``` bash
export GOOS=darwin 
export GOARCH=amd64
go build -o ./bin/ava .
```

#### Using Makefile

``` bash
make build
```

### Usage

To use AVA CLI, run the following command:

```bash
./bin/ava help 
```

```bash
./bin/ava visit 'https://ava.ufms.br/course/view.php?id=xxxx' -u 'username' -p 'password'
```