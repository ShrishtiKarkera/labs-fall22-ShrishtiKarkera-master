# Programming Assignments of CSE 486/586

### Overview

The goal of these assignments is to develop the skills of designing and implementing distributed protocols over multiple machines. 
They are more about designing and understanding the protocols/systems rather than programming.

Assignment 1 is a simple application of MapReduce. It familiarizes you with the Go language and the distributed coding environment. 

Assignment 2 works on a more complicated protocol—distributed snapshot. It expects you to tackle more challenging designs. 

Assignment 3 and 4 implement Raft, a complex consensus protocol. It expects you to solve difficult problems in distributed systems.

### Suggestions

* Start early. These assignments are difficult. 
* Understand the protocols and work out your design first before coding.
* Code progressively. (Finishing an assignment in one sitting is impossible!)
* Save your progress frequently. Use Git!


### System and Language Requirement

* You can develop your code on any OS, e.g., MacOS, Windows, Linux. Please note that TAs will grade your assignments on Ubuntu. 
* The assignments are written in Go. The tests are known to work with Go v1.13 and above. 
* Git is required for assignment submission.

### Tools and IDEs
<p>
 There are some useful tools in the Go ecosystem, e.g., 
 <a href="https://golang.org/cmd/gofmt/">Go fmt</a>, <a href="https://golang.org/cmd/vet/">Go vet</a>, and <a href="https://github.com/golang/lint">Golint</a>. 
 These tools could make your coding easier, but you do not have to use them. 
</p>

<p>
For those who are used to Emacs and Vim, there are some resources for Go development, e.g., <a href="https://github.com/dominikh/go-mode.el">go_in_emacs</a> (additional information available <a href="http://dominik.honnef.co/posts/2013/03/emacs-go-1/">here</a>) and <a href="https://github.com/fatih/vim-go">go_in_vim</a> (additional resources <a href="http://farazdagi.com/blog/2015/vim-as-golang-ide/">here</a>).
</p>

<p>
For those who are used to Sublime, there are some useful Sublime packages for Go development: <a href="https://github.com/DisposaBoy/GoSublime">GoSublime</a> and <a href="https://github.com/golang/sublime-build">Sublime-Build</a>, and <a href="https://atom.io/packages/go-plus">Go-Plus</a> (walkthrough and additional info <a href="https://rominirani.com/setup-go-development-environment-with-atom-editor-a87a12366fcf#.v49dtbadi">here</a>).
</p>

<p>
<a href="https://www.jetbrains.com/go/">JetBrains Goland</a> is also a good option. It's a useful, full-featured Go IDE. You can get a free educational account with your UB email address. 
</p>

### Coding Style

<p> Good coding style is always important, and sometimes necessary, e.g., in large collaborative projects. 
Your code should have proper indentation, descriptive comments,
and a comment header at the beginning of each file, which includes
your name, student id, and a description of the file. 
A good coding style is always consistent, e.g., the same format for all comments throughout your code. 
We do not grade your code based on the style, you earn full credits as along as the code passes all tests, 
but you should prepare yourself (starting from now) for future real-world projects.
</p>

<p>It is recommended to use the standard tools <tt>gofmt</tt> and <tt>go
vet</tt>. You can also use the <a
href="https://github.com/qiniu/checkstyle">Go Checkstyle</a> tool for
advice on how to improve your code's style. It would also be advisable to
produce code that complies with <a
href="https://github.com/golang/lint">Golint</a> where possible. </p>

### Git

<p> Version control is necessary for developing large collaborative projects while working as a team. 
These assignments are individual, but you will be familiarized with the basic functions of Git. 
Please read this <a href="https://git-scm.com/docs/gittutorial">Git Tutorial</a>.</p>

<p>The basic Git workflow in the shell (assuming you already have a repo set up):</br>
<ul>
<li>git pull</li>
<li>do some work</li>
<li>git status (shows what has changed)</li>
<li>git add <i>all files you want to commit</i></li>
<li>git commit -m "brief message on your update"</li>
<li>git push</li>
</ul>
</p>

<p> All programming assignments require Git for submission.</p> 
<p> We use <a href="https://github.com/">Github</a> for distributing and collecting your assignments. (You need to create a Github account if you have not done so.) 
You now should have your working copy of the assignments on Github, named labs-fall22-[your_github_username], by joining the Github classroom. It should be private. Never make it public or share it with anyone else; otherwise, it is a violation of academic integrity. 
To work on the assignments on your local machine, you need to clone the repository from Github to your machine. 
Normally, you only need to clone the repository once.</p>

```bash
$ git clone https://github.com/Distributed-Systems-at-Buffalo/labs-fall22-[username].git 586
$ cd 586
$ ls
assignment1-1  assignment1-2  assignment1-3  assignment2  assignment3  assignment4  README.md
$ 
```

Now, you have everything you need for doing all assignments, i.e., instructions and starter code. 
Git allows you to keep track of and save the changes you make to the code. (This is why it's called version control.) 
For example, to checkpoint your progress, you can <emph>commit</emph> your changes by running:

```bash
$ git commit -am 'partial solution to main 1-1'
$ 
```
Commit is to package your recent changes on your local machine, snapshot it with a version number, and allow you to revert some changes or go back to 
previous version (commit) of your code. At this point, your code is mostly safe, e.g., accidentally deleting some lines/files is not the end of the world 
(ctl+z cannot save you in this case but Git can allow you to restore the code to some recent commit.) 
You should do this regularly!  

Commit is mostly safe because your local machine could fail/crash. You should always pair commit with push. Push is to upload your commits to Github (a remote fault-tolerant machine). 
You can _push_ your changes to Github after you commit with:

```bash
$ git push origin master
$ 
```

After push, you should be able to see your recently pushed commits on Github website.
Please let us know that you've gotten this far in the assignment, by pushing a tag to Github. A tag is a descriptive flag attached to a specific commit. 

```bash
$ git tag -a -m "i got git and cloned the assignments" gotgit
$ git push origin gotgit
$
```

As you complete parts of the assignments (and begin future assignments) we'll ask you push tags. You should also be committing and pushing your progress regularly.
We grade your assignments by pulling your tagged commits from Github. So please push and tag your submission commit! We do not grade your unpushed commits on your local machine. 

### Stepping into Assignment 1-1

Now it's time to go to the [assignment 1-1](assignment1-1) folder to begin your adventure!

### Acknowledgements
<p>Some of the assignments are adapted from MIT's 6.824 course. Thanks to Frans Kaashoek, Robert Morris, and Nickolai Zeldovich for their support.</p>
