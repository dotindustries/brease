# brease: business rules engine as a service

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/dotindustries/brease/blob/master/LICENSE)
![Discord Shield](https://discordapp.com/api/guilds/1010058789812195408/widget.png?style=shield)

brease is an open-source monorepo that provides a robust and extensible business rules engine for various applications. It offers a comprehensive API backend server written in Go, along with TypeScript libraries for core functionality and integrations with popular front-end frameworks.

## Features

- **brease Server**: The `api` folder contains a powerful and efficient backend server implemented in Go. It serves as the backbone for the entire business rules engine, offering high performance and scalability.

- **@brease/core**: This TypeScript library, located in the `packages/core` directory, provides framework-agnostic basic functionality. It includes essential utilities and helpers to support the implementation of business rules within your application.

- **@brease/react**: The `packages/react` directory contains the `@brease/react` TypeScript library, which offers a React provider and a collection of simple React hooks. These components enable seamless integration of the brease business rules engine with your React applications.

- **@brease/vue**: Located in the `packages/vue` directory, the `@brease/vue` package provides Vue 3 composables for easy integration of the brease business rules engine into your Vue applications. It simplifies the process of incorporating dynamic business rules and logic into your Vue components.

## Getting Started

To use brease in your project, follow these steps:

1. Clone the repository:

   ```bash
   git clone https://github.com/dotindustries/brease.git
   ```

2. Install dependencies:

   ```bash
   cd brease
   pnpm install
   ```

3. Build the Go API backend server:

   ```bash
   cd apps/api
   go build
   ```

4. Start the API server:

   ```bash
   ./api
   ```

5. Install the desired TypeScript libraries (e.g., @brease/core, @brease/react, or @brease/vue) in your project:

   ```bash
   npm install @brease/core
   ```

   or

   ```bash
   npm install @brease/react
   ```

   or

   ```bash
   npm install @brease/vue
   ```

For more detailed instructions, please refer to the individual package directories.

## Documentation

For comprehensive documentation on how to use the brease business rules engine, refer to the [official documentation](https://docs.brease.run). It provides detailed explanations, tutorials, and examples to help you integrate and leverage the engine's features effectively.

## Examples

The `examples` directory contains several demo applications showcasing the usage of brease with different front-end frameworks. Explore these examples to gain a better understanding of how to implement the business rules engine within your own projects.

## Contributing

We welcome contributions from the community to enhance and expand the functionality of brease. If you would like to contribute, please review our [contribution guidelines](CONTRIBUTING.md) for instructions on how to get started.

## License

brease is licensed under the [MIT License](LICENSE).

## Contact

If you have any questions, suggestions, or feedback, feel free to reach out to us at dev@dot.industries.

## Cloud-Hosted Version (Coming Soon)

We are excited to announce that we are actively developing a cloud-hosted version of brease. This version will provide a scalable and managed environment for running your business rules and logic.

### Key Features

- **Scalability**: Seamlessly scale your business rules engine based on your application's needs. Handle large volumes of rules and requests with ease.

- **Reliability**: Benefit from a robust infrastructure that ensures high availability and minimal downtime for your business rules engine.

- **Ease of Use**: Enjoy a user-friendly interface and simplified management process, allowing you to focus on building and managing your rules without worrying about infrastructure or updates.

Stay tuned for updates as we finalize the development and launch of the cloud-hosted version. We are committed to delivering a powerful and efficient solution for businesses of all sizes.

## Sign up for Updates

To receive the latest news and updates about the cloud-hosted version of Brease, sign up for our mailing list [here](https://brease.run/signup). We'll keep you informed about the release date, beta testing opportunities, and other important announcements.

## Feedback and Suggestions

We value your input and would love to hear your thoughts on the cloud-hosted version. If you have any specific features or requirements you'd like to see implemented, or if you have any questions or suggestions, please reach out to us at dev@dot.industries.

We appreciate your support and look forward to providing you with an exceptional cloud-hosted experience for your business.

---

By contributing to this project, you agree to abide by the [code of conduct](CODE_OF_CONDUCT.md). Let's build a vibrant and inclusive community together!
