# **Blockie - Block Analyser for Ignite**

Blockie is a terminal-based block explorer for the Ignite blockchain that allows developers to view detailed information about each block in real-time as they are processed. This powerful tool provides insights directly in the terminal, including metrics such as block processing time, block height, specific block headers, their structure, and the complete JSON data of the blocks.

## **Features**

- **Real-Time Updates**: See live details of blocks as they are added to the blockchain.
- **Detailed Insights**: Access to block time, height, headers, and complete JSON structure.
- **Developer-Focused**: A tool designed with blockchain developers in mind for enhanced insight into the blockchain they are working on.

## Prerequisites

Have a blockchain running. Favorably with [Ignite CLI](https://github.com/ignite/cli) and currently only default ports (26657)

```bash
ignite chain serve
```

## **Installation**

To install Blockie, ensure you have Ignite CLI installed and follow these steps:

1. Run the following command to add the app to your global configuration:

```bash
ignite app install -g github.com/toschdev/blockie
```

2. You can now use the **`ignite blockie start`** command to launch.

## **Development Workflow**

When developing with Blockie, use this simple loop for an efficient workflow:

1. Clone the repository
2. Install your local cloned repository
```bash
ignite app install -g $(pwd)
```
3. Make changes to the plugin code as needed.
4. Execute **`ignite blockie start`** to recompile the app and test your changes.
5. If your plugins become corrupted, you can remove them by editing your **`igniteapps.yml`** file:

```
nano ~/.ignite/apps/igniteapps.yml # Then remove the line with the app
```

## **Configuration**

To fine-tune the block processing and give each block more time for examination, you can adjust the following settings in your **`config.yml`** file under the **`validators`** section:

```yaml
yamlCopy code
validators:
  - name: alice
    bonded: 100000000stake
    config:
      consensus:
        timeout_commit: '6s'

```

## **Contributing**

We welcome contributions from the community! If you'd like to contribute to Blockie, please fork the repository, make your changes, and submit a pull request.

## **Support**

If you encounter any issues or have questions, please file an issue in the repository's issue tracker.

## **License**

Blockie is released under the Apache2 License.

![Blockie](cubie_ignite.png "Blockie")