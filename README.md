# git-controller

[![REUSE status](https://api.reuse.software/badge/github.com/open-component-model/git-controller)](https://api.reuse.software/info/github.com/open-component-model/git-controller)

## Description

This is the main repository for `git-controller`. The `git-controller` is designed to enable the automated deployment of
software using the [Open Component Model](https://ocm.software) and Flux.

## Functionality

`git-controller` provides the following two main functionalities.

### Syncing

The `Sync` API objects takes a snapshot's output and pushes it into a specific repository using some pre-configured
commit information.

A sample yaml for running a sync operation can look something like this:

```yaml
apiVersion: delivery.ocm.software/v1alpha1
kind: Sync
metadata:
  name: git-sample
  namespace: ocm-system
spec:
  commitTemplate:
    email: e2e-tester@gitea.com
    message: "Update made from git-controller"
    name: Testy McTestface
  interval: 10m0s
  subPath: .
  snapshotRef:
    name: podinfo-deployment-t5bhemw
  repositoryRef:
    name: new-repository-2 # The name of the Repository object that contains access information to the repository.
  automaticPullRequestCreation: true
```

The `repositoryRef` information contains a link to the Repository object explained in section [Repository Management](#repository-management).
That object contains information on how to access the repository and what credentials to use.

Setting `automaticPullRequestCreation: true` will create a Pull Request of the changes. If no branch information is
provided the changes are created from a random generated branch to `main`. The pull request can further be fine-tuned
with the following details:

```yaml
pullRequestTemplate:
  title: This is the title that will be used.
  description: Contains more information about the Pull Request.
  base: feature-branch-1
```

### Repository Management

The Repository object manages git repositories for supported providers. At the moment of this writing the following
providers are supported:

- GitHub
- Gitlab
- Gitea

The main objective of this object is to create a Repository. Along that, it also sets up some branch protection rules.
Branch protection rules are used during the Validation processes in the MPAS environment.

A sample repository could look something like this:

```yaml
apiVersion: mpas.ocm.software/v1alpha1
kind: Repository
metadata:
  name: new-repository-2
  namespace: mpas-system
spec:
  isOrganization: false
  visibility: public
  credentials:
    secretRef:
      name: git-sync-secret
  interval: 10m
  owner: open-component-model
  provider: github
  existingRepositoryPolicy: adopt
```

There are several things here to unpack. First, is the organization. This needs to be set in case the owner of the
future repository is an organization. By default, this is set to `true`. Here, we use `false` for testing purposes.

The second is `visibility`. This can be switched from `public` to `private`.

`credentials` are self-explanatory. Either a token or SSH credentials are supported.

Provider is between `github`, `gitlab` or `gitea`. And finally, we use `existingRepositoryPolicy` to decide what to do
in case the repository already exists. `adopt` will use the repository as is. Not setting it will fail the process if
the repository already exists.

## Testing

`git-controller` usually doesn't run on its own. Since most of its features require a Snapshot to be present. And a
Snapshot is created by the ocm-controller. However, if testing only involves the `Repositroy` object, make sure that a
certificate TLS secret existing in the cluster with the name `ocm-registry-tls-certs`. This can be generated with
`mkcert` or by the test cluster prime script under [ocm-controller](https://github.com/open-component-model/ocm-controller/blob/4109172a978c6e07733870eda85dc2b0029e8e8b/hack/prime_test_cluster.sh).

`git-controller` has a `Tiltfile` which can be used for rapid development. [tilt](https://tilt.dev/) is a convenient
little tool to spin up a controller and do some extra setup in the process conditionally. It will also keep updating
the environment via a process that is called [control loop](https://docs.tilt.dev/controlloop.html); it's similar to
a controller's reconcile loop.

To get started simple run `tilt up` then hit `<space>` to enter Tilt's ui. You should see git-controller starting up.

## Licensing

Copyright 2025 SAP SE or an SAP affiliate company and Open Component Model contributors.
Please see our [LICENSE](LICENSE) for copyright and license information.
Detailed information including third-party components and their licensing/copyright information is available [via the REUSE tool](https://api.reuse.software/info/github.com/open-component-model/ocm-controller).
