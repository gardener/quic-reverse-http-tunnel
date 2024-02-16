# <repo name>

[![reuse compliant](https://reuse.software/badge/reuse-compliant.svg)](https://reuse.software/)

## How to use this repository template

This template repository can be used to seed new git repositories in the gardener github organisation.

- [Create the new repository](https://docs.github.com/en/free-pro-team@latest/github/creating-cloning-and-archiving-repositories/creating-a-repository-from-a-template)
  based on this template repository
- Replacing placeholders:
  - In file `.reuse/dep5` replace placeholder `<repo name>` with the name of your new repository.
  - In file `CODEOWNERS` replace `<repo name>` and `<maintainer team>`. Use the name of the github team in [gardener teams](https://github.com/orgs/gardener/teams) defining maintainers of the new repository.
- Set the repository description in the "About" section of your repository
- Describe the new component in additional sections in this `README.md`
- Ask the [Owner of the gardener github organisation](https://github.com/orgs/gardener/people?query=role%3Aowner)
  - to double-check the initial content of this repository
  - to create the maintainer team for this new repository
  - to make this repository public
  - protect at least the master branch requiring mandatory code review by the maintainers defined in CODEOWNERS
  - grant admin permission to the maintainers team of the new repository defined in CODEOWNERS

## Maintain copyright and license information
By default all source code files are under `Apache 2.0` and all markdown files are under `Creative Commons` license.

When creating new source code files the license and copyright information should be provided using corresponding SPDX headers.

Example for go source code files (replace `<year>` with the current year)
```
/*
 * SPDX-FileCopyrightText: <year> SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */
```

### Third-party source code

If you copy third-party code into this repository or fork a repository, you must keep the license and copyright information (usually defined in the header of the file).

In addition you should adapt the `.reuse/dep5` file and assign the correct copyright and license information.

**Example `dep5` file if you copy source code into your repository:**
```
Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: Gardener <repo name>
Upstream-Contact: The Gardener project <gardener@googlegroups.com>
Source: https://github.com/gardener/<repo name>

# --------------------------------------------------
# source code

Files: *
Copyright: 2017-2024 SAP SE or an SAP affiliate company and Gardener contributors
License: Apache-2.0

# --------------------------------------------------
# documentation

Files: *.md
Copyright: 2017-2024 SAP SE or an SAP affiliate company and Gardener contributors
License: CC-BY-4.0

# --------------------------------------------------
# third-party

# --- copied source code ---
Files: pkg/utils/validation/kubernetes/core/*
Copyright: 2014 The Kubernetes Authors.
License: Apache-2.0
```
**Example `dep5` file if you have forked a repository:**
```
Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: Gardener fork of kubernetes/autoscaler
Upstream-Contact: The Gardener project <gardener@googlegroups.com>
Source: https://github.com/gardener/autoscaler
Comment: This is a fork of kubernetes/autoscaler (https://github.com/kubernetes/autoscaler)

# --------------------------------------------------
# source code

Files: *
Copyright: 2016-2018 The Kubernetes Authors.
License: Apache-2.0

Files: .ci/*
Copyright: 2024 SAP SE or an SAP affiliate company and Gardener contributors
License: Apache-2.0
```

#### Modifications
In case you modify copied/forked source code you must state this in the header via the following text:

**Modifications Copyright <year> SAP SE or an SAP affiliate company and Gardener contributors**


### Get your reuse badge
To get your project reuse compliant you should register it [here](https://api.reuse.software/register) using your SAP email address. After confirming your email, an inital reuse check is done by the reuse API.

To add the badge to your project's `README.md` file, use the snipped provided by the reuse API.
