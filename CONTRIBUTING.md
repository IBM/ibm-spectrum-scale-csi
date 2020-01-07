[go_effectivego]: https://golang.org/doc/effective_go.html
[go_gettingstarted]: https://github.com/golang/go/wiki#getting-started-with-go
[go_goreturns]: https://github.com/sqs/goreturns
[go_golangci-lint]: https://github.com/golangci/golangci-lint#install
[go_codecomments]: https://github.com/golang/go/wiki/CodeReviewComments

# Contributing

## Before You Submit a Pull Request

1. Follow [coding practices](#coding-practices)
2. [Verify](#run-local-verifications) your pull request
3. [Sign your work](#sign-your-work-for-submittal) (required)

### Coding Practices

* Go 
    - get started with [go-lang][go_gettingstarted]
    - practice [Effective Go][go_effectivego]
    - format code with [goreturns][go_goreturns]
    - keep in mind [common review comments][go_codecomments]

### Run Local Verifications
You can run these verifications locally to predict the pass or fail of continuous integration.

* ```goreturns -d ./```
* ```golangci-lint run```
* ```go test ./tests/sanity```

#### Hint: configure your editor to use goreturns and golangci-lint
```bash
golangci-lint run --fast --enable=golint,gocritic
```
>Note: it is recommended to use golint and gocritic, even though they aren't enforced.

```bash
goreturns
```

### Sign your work for submittal

The sign-off is a simple line at the end of the explanation for the patch. Your signature certifies that you wrote the patch or otherwise have the right to pass it on as an open-source patch. The rules are pretty simple: if you can certify the below (from developercertificate.org):

Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
1 Letterman Drive
Suite D4700
San Francisco, CA, 94129

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.

Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
Then you just add a line to every git commit message:

Signed-off-by: Joe Smith <joe.smith@email.com>
Use your real name (sorry, no pseudonyms or anonymous contributions.)

If you set your user.name and user.email git configs, you can sign your commit automatically with git commit -s.

Note: If your git config information is set properly then viewing the git log information for your commit will look something like this:

Author: Joe Smith <joe.smith@email.com>
Date:   Thu Feb 2 11:41:15 2018 -0800

    docs: Update README

    Signed-off-by: Joe Smith <joe.smith@email.com>
Notice the Author and Signed-off-by lines match. If they don't your PR will be rejected.