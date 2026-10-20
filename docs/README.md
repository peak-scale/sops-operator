# Documentation

This directory contains the Hugo site using the same Doks layout as
`schedule-action-controller`. Content lives in `content/docs/`.

From the repository root, enter `devbox shell`, then run:

```shell
task docs:serve    # http://localhost:1313
task docs:generate # regenerate the API reference from Helm CRDs
task docs:check    # check generated content and links at root and subpath URLs
task docs:build    # build into dist/docs
```

The npm dependencies are locked in `package-lock.json` and installed with
`npm ci --ignore-scripts`. Builds default to `https://sops.addons.projectcapsule.dev/`;
CI uses the URL configured in GitHub Pages. Set `DOCS_BASE_URL` when building for
another host or subpath.
