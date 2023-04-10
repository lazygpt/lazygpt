# LazyGPT 🤖💬🚀

Welcome to **LazyGPT**! This extensible AI agent leverages advanced language
models and plugins to make planning and implementation a breeze. 🌬️🌟

LazyGPT is an autonomous agent that utilizes GPT or other language models to
develop plans and implement them. It can run on the CLI or serve a web UI,
and everything from the language model to the various commands exposed to the
model are implemented with plugins.

## Features 🌈

- 🗣️ Uses GPT and other language models for intelligent planning
- 🔌 Highly modular and extensible with plugins

## Installation 💻

To install LazyGPT, simply clone the repository and build the project:

1. Clone the repository:
```bash
git clone https://github.com/lazygpt/lazygpt.git
````
2. Change into the project directory:
```bash
cd lazygpt
```
3. Build the project in a container:
```
make build
```

## Usage 🚀

LazyGPT can be used either in chat mode or by starting a web server:

### Chat Mode 💬

To interact with LazyGPT in chat mode, run the following command:

```bash
dist/lazygpt chat
```

### Web Server Mode 🌐

To start LazyGPT's web server and serve the web UI, run the following command:

`./lazygpt serve`

## Plugins 🧩

LazyGPT supports multiple interfaces for plugins, allowing you to extend its
functionality. To create a plugin, simply implement the desired interface and
register it with LazyGPT.

## Contributing 🤝

Contributions to LazyGPT are welcome! To contribute, please fork the
repository, make your changes, and submit a pull request. Be sure to follow
the code style and provide tests for any new features or bug fixes.

## License 📄

LazyGPT is released under the [MIT License](LICENSE).
