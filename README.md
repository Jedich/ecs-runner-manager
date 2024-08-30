# aws-ecs-runner-manager

This is a simple tool built on **Golang** that provides automatic management for Self-hosted GitHub Actions runners on Fargate/EC2 type ECS, designed to be as easy to integrate as possible.

## Main Features

- **Runner Manager** that manages the lifecycle of the runners.
  - Supports metrics fetching over controlled runners via Prometheus exporter
  - Deployed on ECS as a service or task
- **Monitoring agent** that can be deployed to monitor and control the runner's and runner controllers state, metrics, logs, and more. Contains a web server and a dashboard client built on React.
- **Runner configuration** that will be used to run the GitHub Actions jobs. Can be configured manually or via Monitoring agent.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
