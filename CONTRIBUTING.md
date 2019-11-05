# Contributing

Work on the `ibm-spectrum-scale-csi-operator` should always be performed in a forked copy of the repo, incorporated into the main project using a pull request. 


When adding features the following process should be followed:

0. Make sure you've been [approved to contribute code](sign-your-work-for-submittal)
1. Write a [Mini Design Document](https://github.com/IBM/ibm-spectrum-scale-csi-operator/issues/new?assignees=&labels=Epic&template=mini-design-document.md&title=%5BMDD%5D+New+Feature) for your feature.
2. On your [fork](#forking-the-repo), create a branch for your feature
3. Create a [pull request](https://github.com/IBM/ibm-spectrum-scale-csi-operator/compare) with your features.


## Forking the Repo

The following is a quick guide to forking IBM Spectrum Scale CSI Operator.

### GUI

1. Select the fork option on github:

<img width="1087" alt="Screen Shot 2019-11-05 at 4 49 16 PM" src="https://user-images.githubusercontent.com/1195452/68249220-5716fd80-ffec-11e9-8b3c-f0c70564f055.png">

2. Select your user and create the fork.

<img width="444" alt="image" src="https://user-images.githubusercontent.com/1195452/68249628-2f746500-ffed-11e9-9b2c-f27e9dfd418e.png">

### Dev Environment

To pull your fork into your environment, run the following commands:

``` bash
# Set up some helpful variables
export GOPATH=<your-go-path>
export USERNAME=<github-username>
export IBM_DIR="$GOPATH/src/github.com/IBM"
export OPERATOR_DIR="$IBM_DIR/ibm-spectrum-scale-csi-operator"

# Ensure the dir is present then clone.
mkdir -p ${IBM_DIR}
cd ${IBM_DIR}
git clone git@github.com/${USERNAME}/ibm-spectrum-scale-csi-operator.git
git remote add upstream git@github.com:IBM/ibm-spectrum-scale-csi-operator.git
```

At this point, you should have a `origin` remote pointing to your forked repo, and a `upstream` remote pointing to the main repository.

Before branching remember the following:
1. Start your branches from `dev`.
2. Sync your `dev` branch using `git pull upstream dev`
 * Be sure to push the merged copy to your forked repo using `git push origin dev`


## Sign your work for submittal

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
