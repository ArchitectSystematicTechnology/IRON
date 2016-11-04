# Creating Docker images out of Lambda functions
Docker images created by running the `create-function` subcommand on a Lambda function are ready to execute. 

You can convert any Lambda function of type nodejs (0.10), python (2.7) and Java (8) into an Iron Function compatible Docker Image as follows:
```bash
fnctl lambda create-function <name> <runtime> <handler> <files...>
```

where:
* name: the name of the created docker image which should have the format <username>/<image-name>
* runtime: any of the following `nodejs`, `python2.7` or `java8`
* handler: a handler takes a different form per runtime
    * java8: `<namespace>.<class>::<handler>`
    * python2.7:  `<filename>.<handler>`
    * nodejs: `<filename>.<handler>`
* file: the files to be converted, however for java8 only one file of type `jar` is allowed.

e.g:
```bash
fnctl lambda create-function irontest/node-exec:1 nodejs node_exec.handler node_exec.js
```

