[
  {
    version: 2,
    workflows: {
      version: 2,

      'build-and-test': {
        jobs: ['test'],
      },
    },

    jobs: {
      test: {
        docker: [{ image: 'circleci/golang:1.13' }],
        resource_class: 'small',

        environment: {
          GOFLAGS: '-mod=readonly -p=8',  // Go on Circle thinks 32 CPUs are available, but there aren't.
        },

        steps: [
          'checkout',
          { run: 'go test -race ./...' },
          {
            run: {
              name: 'Sync YAML',
              command: 'go run -race . .circleci/config.jsonnet .circleci/config.yml',
            },
          },
          {
            run: {
              name: 'Confirm no diff after syncing YAML',
              command: 'test -z "$(git status --porcelain)" || (echo "Changes detected after running make generate"; git status; git --no-pager diff; false)',
            },
          },
        ],
      },
    },
  },
]
