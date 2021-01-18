**Based on https://github.com/git-time-metric/gtm**

<p align="center">
    <img src="./readme/logo.svg" width="256" height="256" alt="logo">
</p>

### Seamless time tracking for all your Git projects
![Release](https://github.com/DEVELOPEST/gtm-core/workflows/Release/badge.svg?branch=master)
![Develop](https://github.com/DEVELOPEST/gtm-core/workflows/Develop/badge.svg?branch=develop)

## Installation
See wiki: https://github.com/DEVELOPEST/gtm-core/wiki/Installation

##### $ gtm report -last-month
<div><img src="https://cloud.githubusercontent.com/assets/630550/21582250/8a03f9dc-d015-11e6-8f77-548ef7314bf7.png"></div>

##### $ gtm report -last-month -format summary
<div><img src="https://cloud.githubusercontent.com/assets/630550/21582252/8f85b738-d015-11e6-8c70-beed7e7b3254.png"></div>

##### $ gtm report -last-month -format timeline-hours
<div><img src="https://cloud.githubusercontent.com/assets/630550/21582253/91f6226e-d015-11e6-897c-6042111e6a6a.png"></div> </br>

GTM is automatic, seamless and lightweight.  There is no need to remember to start and stop timers.  It runs on occasion to capture activity triggered by your editor.  The time metrics are stored locally with the git repository as [Git notes](https://git-scm.com/docs/git-notes) and can be pushed to the remote repository.

### Plugins

Simply install a plugin for your favorite editor and the GTM command line utility to start tracking your time now.

### Initialize a project for time tracking

<pre>$ cd /my/project/dir
$ gtm init

Git Time Metric initialized for /my/project/dir

         post-commit: gtm commit --yes
            pre-push: git push origin refs/notes/gtm-data --no-verify
      alias.fetchgtm: fetch origin refs/notes/gtm-data:refs/notes/gtm-data
       alias.pushgtm: push origin refs/notes/gtm-data
   notes.rewriteMode: concatenate
    notes.rewriteRef: refs/notes/gtm-data
       add fetch ref: +refs/notes/gtm-data:refs/notes/gtm-data
            terminal: true
          .gitignore: /.gtm/
                tags: tag1, tag2
</pre>

### Edit some files in your project

Check your progress with `gtm status`.  

<pre>$ gtm status

       20m 40s  53% [m] plugin/gtm.vim
       18m  5s  46% [r] Terminal
           15s   1% [m] .gitignore
       39m  0s          <b>gtm-vim-plugin</b> </pre>

### Commit your work

When you are ready, commit your work like you usually do.  GTM will automatically save the time spent associated with your commit. To check the time of the last commit type `gtm report`.
<pre>$ gtm report

7129f00 <b>Remove post processing of status</b>
Fri Sep 09 20:45:03 2016 -0500 <b>gtm-vim-plugin</b> Michael Schenk

       20m 40s  53% [m] plugin/gtm.vim
       18m  5s  46% [r] Terminal
           15s   1% [m] .gitignore
       39m  0s          <b>gtm-vim-plugin</b> </pre>

### Optionally save time in the remote Git repository

GTM provides [git aliases](https://git-scm.com/book/en/v2/Git-Basics-Git-Aliases) to make this easy.  It defaults to origin for the remote repository.

Time data can be saved to the remote repository by pushing.
<pre>$ git pushgtm </pre>

Time data can be retrieved from the remote repository by fetching.
<pre>$ git fetchgtm </pre>

### Getting Help

For help from the command line type `gtm --help` and `gtm <subcommand> --help`.

For additional help please consult the [Wiki](https://github.com/DEVELOPEST/gtm-core/wiki).

# Contributing
If you find a bug or have an idea for a new feature please feel free to file new issues and submits PRs.  
In a particular if there isn't a plugin for your favorite editor, go ahead and create one!

For more detail on how to write plugins, check out the [Wiki](https://github.com/DEVELOPEST/gtm-core/wiki/Editor-Plugins).

# Building from source

### Install go
You can follow any go installation tutorial, one for Ubuntu for example: https://medium.com/golang-basics/installing-go-on-ubuntu-b443a8f0eb55

### Install apt dependencies
```bash
sudo apt-get install libgit2-dev libssh2-1-dev libssl-dev cmake
```

### Clone gtm-core
```bash
git clone https://github.com/DEVELOPEST/gtm-core.git
mv gtm-core $GOPATH/src/github.com/DEVELOPEST/
cd $GOPATH/src/github.com/DEVELOPEST/gtm-core
git submodule update --init  # Install vendor dependecies
```

### Install dependencies
```bash
go get -d github.com/Masterminds/sprig
go get -d github.com/libgit2/git2go
cd $GOPATH/src/github.com/libgit2/git2go
git submodule update --init  # get libgit2
cd vendor/libgit2
mkdir build && cd build
cmake ..
make
sudo make install
cd ../../..
make install-static
```

### Build
```bash
cd $GOPATH/src/github.com/DEVELOPEST/gtm-core
mkdir build
go build -o build/ ./...
```

# Support

To report a bug, please submit an issue on the [GitHub Page](https://github.com/DEVELOPEST/gtm-core/issues)

Consult the [Wiki](https://github.com/DEVELOPEST/gtm-core/wiki) for more information.
